sqssh
=====

Install
-------

`go get github.com/squarescale/sqssh`

Usage
-----

```
$ cat sqssh.yaml
---
hosts:
  - name: worker
    user: core
    jump: bastion
    filters:
      - Name: tag:Name
        Values: Worker
      - Name: instance-state-name
        Values: running
  - name: bastion
    user: core
    public: true
    filters:
      - Name: tag:Name
        Values: Bastion
      - Name: instance-state-name
        Values: running

$ sqssh user@bastion
core@ip-10-0-253-97 ~ $ exit

$ # through a jumphost defined in configuration
$ sqssh user@worker
core@ip-10-0-26-169 ~ $ exit

$ # use any options like you would usually do
$ sqssh -Nf -L8500:localhost:8500 user@worker
$ curl -i localhost:8500 |head -n1
HTTP/1.1 200 OK
```

The tool leaves the `user@host` as-is in the command line, and will prepend `-o Hostname ec2-whatever` to ssh args'. This means that ssh will read config for the host you specify on the command-line. This allows to configure any named host that can change in your cluster.