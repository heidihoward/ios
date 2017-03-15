FROM golang
ADD . /go/src/github.com/heidi-ann/ios
EXPOSE 8080
RUN go get github.com/heidi-ann/ios/...
ENTRYPOINT ["server"]
