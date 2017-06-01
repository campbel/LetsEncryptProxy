FROM golang:1.8.3-alpine
ADD . /go/src/github.com/campbel/LetsEncryptProxy
RUN go install github.com/campbel/LetsEncryptProxy
CMD LetsEncryptProxy