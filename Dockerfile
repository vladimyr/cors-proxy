FROM golang:latest
LABEL name "cors-proxy"
ADD . /go/src/github.com/vladimyr/cors-proxy
RUN go install github.com/vladimyr/cors-proxy
EXPOSE 3000
CMD ["/go/bin/cors-proxy", "-p", "3000"]
