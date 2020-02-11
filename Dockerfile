# build image
FROM golang:buster AS builder

WORKDIR $GOPATH/src/github.com/rwn3120/wimp

COPY * ./

RUN CGO_ENABLED=0 go build -o /go/bin/wimp

# image
FROM alpine:3

ARG PORT=8080

COPY --from=builder /go/bin/wimp /bin/

EXPOSE ${PORT}

ENTRYPOINT ["/bin/wimp"]
