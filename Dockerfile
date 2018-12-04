FROM golang:1.11 AS builder
LABEL description="sqssh docker image"
LABEL version="0.1"
LABEL maintainer "engineering@squarescale.com"
ENV USER sqssh
RUN adduser --disabled-password --gecos ''  $USER
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o /go/bin/sqssh

FROM alpine
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/sqssh /go/bin/sqssh
ENV USER sqssh
USER $USER
ADD sqssh.yml /home/$USER/.config/
ENTRYPOINT ["/go/bin/sqssh"]
