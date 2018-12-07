package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/awsutil"
	"errors"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docopt/docopt-go"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
	"syscall"
)

const ssh_usage = `
	usage: ssh [-46AaCfGgKkNnqsTVXxYy] [-v...] [-M...] [-t...] [-B <bind_interface>]
   [-b <bind_address>] [-c <cipher_spec>] [-D <dynamic>]
   [-E <log_file>] [-e <escape_char>] [-F <configfile>] [-I <pkcs11>]
   [-i <identity_file>...] [-J <jumpspec>] [-L <address>...]
   [-l <login_name>] [-m <mac_spec>] [-O <ctl_cmd>] [-o <option>] [-p <port>]
   [-Q <query_option>] [-R <address>...] [-S <ctl_path>] [-W <host:port>]
   [-w <tunspec>] DESTINATION [COMMAND...]
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

type SshCommand struct {
	Args []string
	Opts docopt.Opts
}

func NewSshCommand(args []string) (*SshCommand, error) {
	sc := new(SshCommand)
	sc.Args = args

	parser := &docopt.Parser{
		HelpHandler:   docopt.NoHelpHandler,
		SkipHelpFlags: true,
	}

	opts, err := parser.ParseArgs(ssh_usage, sc.Args[1:], "")
	sc.Opts = opts
	log.Debugf("opts: %+v", sc.Opts)
	return sc, err
}

func (s *SshCommand) hostnameFromCommand(destination string) string {
	if strings.Contains(destination, "@") {
		destination = strings.Split(destination, "@")[1]
	}
	return destination
}

func ec2Init() (*ec2.EC2, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		log.Debug(err)
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, errors.New("no AWS credential provided, see: https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/setup-credentials.html")
	}

	ec2svc := ec2.New(sess)
	//ec2svc := ec2.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
	return ec2svc, nil
}

func ec2InstancesRequest(ec2svc *ec2.EC2, name string) (*ec2.DescribeInstancesOutput, error) {
	filters := []*ec2.Filter{
		// builtin filter, the purpose is connecting to hosts, they need to be running
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
	}

	f := &ec2.Filter{
		Name:   aws.String("tag:Name"),
		Values: []*string{aws.String(name)},
	}
	filters = append(filters, f)

	params := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	resp, err := ec2svc.DescribeInstances(params)
	if err != nil {
		fmt.Println("there was an error listing instances in", err.Error())
		return nil, err
	}
	return resp, err

}

func chooseInstanceInteractively(resp *ec2.DescribeInstancesOutput) string {
	hosts := []string{}
	hostname := ""
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			hostname = *inst.PublicDnsName
			if hostname == "" {
				hostname = *inst.PrivateDnsName
			}
			hosts = append(hosts, hostname)

			log.Debugf("dns: %#v", hostname)
		}
	}

	if len(hosts) > 1 && terminal.IsTerminal(syscall.Stdin) {
		prompt := promptui.Select{
			Label: "Which instance",
			Items: hosts,
		}
		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
		} else {
			hostname = result
		}
	}
	if len(hosts) == 1 {
		hostname = hosts[0]
	}

	return hostname
}

func hostnameFromEc2(hostname string) string {
	ec2svc, _ := ec2Init()
	instances, _ := ec2InstancesRequest(ec2svc, hostname)
	return chooseInstanceInteractively(instances)
}

func (s *SshCommand) hostnameFromConfig(host string) string {
	n := viper.GetString("hosts." + host + ".name")
	if n == "" {
		n = host
	}
	return host
}

func (s *SshCommand) hostArg() string {
	h := s.hostnameFromCommand(s.Opts["DESTINATION"].(string))
	h = s.hostnameFromConfig(h)
	h = hostnameFromEc2(h)
	if h != "" {
		return "-o Hostname " + h
	}
	return ""
}

func (s *SshCommand) jumpArg() string {
	h := s.hostnameFromCommand(s.Opts["DESTINATION"].(string))
	j := viper.GetString("hosts." + h + ".jump")
	if j != "" {
		j := s.hostnameFromConfig(j)
		u := viper.GetString("hosts." + j + ".user")
		j = hostnameFromEc2(j)
		if u != "" {
			u = u + "@"
		}
		return "-J " + u + j
	}
	return ""
}

func (s *SshCommand) cmd() []string {
	cmd := []string{"/usr/bin/ssh"}

	h := s.hostArg()
	if h != "" {
		cmd = append(cmd, h)
	}

	j := s.jumpArg()
	if j != "" {
		cmd = append(cmd, j)
	}
	cmd = append(cmd, s.Args[1:]...)
	return cmd
}

func (s *SshCommand) run(args []string) {
	log.Debugf("Executing ssh with: %s", args)
	syscall.Exec("/usr/bin/ssh", args, os.Environ())
}

func main() {
	viper.SetConfigName("sqssh")
	viper.AddConfigPath("$HOME/.config")
	viper.SetEnvPrefix("SQSSH")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	// once logger is configured, we can handle viper err
	if err != nil {
		log.Debug(err)
	}

	sc, err := NewSshCommand(os.Args)
	if err != nil {
		log.Debug("couldn't parse ssh command, executing as-is")
		sc.run(os.Args)
	}

	sc.run(sc.cmd())
}
