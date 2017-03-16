FROM golang
ADD . /go/src/github.com/heidi-ann/ios
EXPOSE 8080 8090
RUN go get github.com/heidi-ann/ios/...
ENTRYPOINT ["ios","-listen-peers=8090","-listen-clients=8080"]
