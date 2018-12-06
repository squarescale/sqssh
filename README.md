sqssh
=====

Install
-------

`go get -u github.com/squarescale/sqssh`

Usage
-----

```
$ cat sqssh.yaml
---
hosts:
  Worker:
    jump: Bastion
  Bastion:
    user: user
  somecoolname:
    name: Bastion

$ sqssh Bastion
user@ip-10-0-253-97 ~ $ exit

$ # through a jumphost defined in configuration
$ sqssh Worker
user@ip-10-0-26-169 ~ $ exit

$ # use any options like you would usually do
$ sqssh -Nf -L8500:localhost:8500 Worker
$ curl -i localhost:8500 |head -n1
HTTP/1.1 200 OK

$ # use ssh config
$ cat <<EOF >>~/.ssh/config
Host Worker
    LocalForward 8500 127.0.0.1:8500
EOF
$ sqssh -Nf Worker
$ curl -i localhost:8500 |head -n1
HTTP/1.1 200 OK


$ # alias names
$ ssh somecoolname
user@ip-10-0-253-97 ~ $ exit
```

The tool leaves the destination (in the form of `user@host` as-is in the command line, and will prepend `-o Hostname ec2-whatever` to ssh arguments. This means that ssh will read config for the host you specify on the command-line. This allows to configure any named host that can change in your cluster.

You can for example have in your ssh_config:

```
Host Worker
    User john
    LocalForward 5000 127.0.0.1:5000
```

Sqssh will build an ssh command that specifies the real address of worker, but that command will in fact read `~/.ssh/config` , and thus, it will use the username `john` to connect and will setup a tunnel on port 5000.
