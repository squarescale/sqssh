package main

import "testing"

func TestNewSshCommand(t *testing.T) {
	// this crashes:
	// sc, err := NewSshCommand([]string{""})

	argstests := [][]string{
		[]string{"ssh"},                   // called as ssh
		[]string{"sqssh"},                 // no args (no host specified)
		[]string{"sqssh", "--help"},       // non existant arg
		[]string{"sqssh", "-i", "id_rsa"}, // no host specified
	}

	for _, tt := range argstests {
		sc, err := NewSshCommand(tt)
		if err.Error() != "" && len(sc.Opts) == 0 {
			t.Logf("working args that should fail: %s", tt)
			t.Fail()
		}
	}

	sc, err := NewSshCommand([]string{"sqssh", "-i"}) // not complete arg
	if err.Error() != "-i requires argument" && len(sc.Opts) == 0 {
		t.Log("working args that should fail")
		t.Log(err)
		t.Fail()
	}

	argstests = [][]string{
		[]string{"ssh", "host"},                                // called ssh
		[]string{"sqssh", "host"},                              // simplest working command
		[]string{"sqssh", "-i", "id_rsa", "host"},              // with an ssh key
		[]string{"sqssh", "host"},                              // with a user specified
		[]string{"sqssh", "-vvv", "host"},                      // a repeatable arg
		[]string{"sqssh", "host", "hostname"},                  // with a command
		[]string{"sqssh", "host", "cat", "file"},               // with a multiple arg command
		[]string{"sqssh", "-L", "5000:localhost:5000", "host"}, // with a tunnel
		[]string{"sqssh", "-L", "5000:localhost:5000",
			"-L", "5001:localhost:5001", "host"}, // with 2 tunnels
	}

	for _, tt := range argstests {
		sc, err := NewSshCommand(tt)
		if sc.Opts["DESTINATION"] != "host" || err != nil {
			t.Logf("non-working args that should work: %s", tt)
			t.Fail()
		}
	}

	sc, err = NewSshCommand([]string{"sqssh", "user@host"}) // specifying a user@host
	if sc.Opts["DESTINATION"] != "user@host" || err != nil {
		t.Log("non-working args that should work")
		t.Fail()
	}
}

func TestHostnameFromCommand(t *testing.T) {
	hostspecs := []string{
		"host",
		"user@host",
		"user @host",
	}
	for _, tt := range hostspecs {
		sc, _ := NewSshCommand([]string{"ssh", "host"})
		hostname := sc.hostnameFromCommand(tt)
		if hostname != "host" {
			t.Log(tt)
			t.Fail()
		}
	}

}

func TestHostnameFromAws(t *testing.T) {
	hostname := hostnameFromEc2("bastion")
	t.Log(hostname)
}

// func (s *SshCommand) hostnameFromAws(hostname string) string
// func (s *SshCommand) hostnameFromConfig(host string) string
// func (s *SshCommand) hostArg() string
// func (s *SshCommand) jumpArg() string
// func (s *SshCommand) cmd() []string
// func (s *SshCommand) run(args []string)
// func main()
