package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docopt/docopt-go"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"syscall"
)

type Filter struct {
	Name   string
	Values string
}

type Host struct {
	Name     string
	Hostname string
	User     string
	Filters  []Filter
	Public   bool
	Jump     string
	Query    string
}

func (h *Host) hostnameFromAws() {
	sess, _ := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	ec2svc := ec2.New(sess)
	//ec2svc := ec2.New(session.New(), aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))

	filters := []*ec2.Filter{}
	for _, filter := range h.Filters {
		f := ec2.Filter{
			Name:   aws.String(filter.Name),
			Values: []*string{aws.String(filter.Values)},
		}
		filters = append(filters, &f)
	}

	params := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	resp, err := ec2svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		log.Fatal(err.Error())
	}

	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if h.Public {
				h.Hostname = *inst.PublicDnsName
			} else {
				h.Hostname = *inst.PrivateDnsName
			}
			break
		}
	}
}

func (h *Host) userHost() string {
	uarg := ""
	if h.User != "" {
		uarg = h.User + "@"
	}
	return uarg + h.Hostname
}

type Config struct {
	Hosts []Host
}

func config() (Config, error) {
	viper.SetConfigName("sqssh")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	var C Config
	err = viper.Unmarshal(&C)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v", err)
	}
	return C, err
}

func findHost(h string, c Config) Host {
	for _, host := range c.Hosts {
		if h == host.Name {
			return host
		}
	}
	return Host{}
}

func main() {
	ssh_usage := `
	usage: ssh [-46AaCfGgKkNnqsTVXxYy] [-v...] [-M...] [-t...] [-B <bind_interface>]
   [-b <bind_address>] [-c <cipher_spec>] [-D <dynamic>]
   [-E <log_file>] [-e <escape_char>] [-F <configfile>] [-I <pkcs11>]
   [-i <identity_file>...] [-J <jumpspec>] [-L <address>...]
   [-l <login_name>] [-m <mac_spec>] [-O <ctl_cmd>] [-o <option>] [-p <port>]
   [-Q <query_option>] [-R <address>...] [-S <ctl_path>] [-W <host:port>]
   [-w <tunspec>] DESTINATION [COMMAND]
options:
    -B <bind_interface>
    -b <bind_address>
    -c <cipher_spec>
    -D <dynamic>
    -E <log_file>
    -e <espace_char>
    -F <configfile>
    -I <pkcs11>
    -i <identity_file>...
    -J <jumpspec>
    -L <address>...
    -l <login_name>
    -m <mac_spec>
    -O <ctl_cmd>
    -o <option>
    -p <port>
    -Q <query_option>
    -R <address>...
    -S <ctl_path>
    -W <host_port>
    -w <tunspec>
`

	parser := &docopt.Parser{HelpHandler: func(error, string) {}}
	opts, _ := parser.ParseArgs(ssh_usage, os.Args[1:], "")

	c, err := config()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	args := modifyArgs(os.Args, c, opts)
	syscall.Exec("/usr/bin/ssh", args, os.Environ())
}

func modifyArgs(args []string, c Config, opts docopt.Opts) []string {
	hostname := ""
	if strings.Contains(opts["DESTINATION"].(string), "@") {
		hostname = strings.Split(opts["DESTINATION"].(string), "@")[1]
	} else {
		hostname = opts["DESTINATION"].(string)
	}
	h := findHost(hostname, c)
	h.hostnameFromAws()

	jarg := ""
	if h.Jump != "" {
		jh := findHost(h.Jump, c)
		jh.hostnameFromAws()
		jarg = jh.userHost()
	}

	var wrappedArgs []string
	for i, arg := range args {
		if i == 1 {
			wrappedArgs = append(wrappedArgs, "-o", "Hostname "+h.Hostname)
			if h.Jump != "" {
				wrappedArgs = append(wrappedArgs, "-J", jarg)
			}
		}
		wrappedArgs = append(wrappedArgs, arg)
	}
	return wrappedArgs
}
