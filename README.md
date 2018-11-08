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

$ SQSSH_HOST=bastion sqssh user@host
core@ip-10-0-253-97 ~ $ exit

$ # through a jumphost defined in configuration
$ SQSSH_HOST=worker sqssh user@host
core@ip-10-0-26-169 ~ $ exit

$ # use any options like you would usually do
$ SQSSH_HOST="worker" sqssh -Nf -L8500:localhost:8500 user@host
$ curl -i localhost:8500 |head -n1
HTTP/1.1 200 OK
```
