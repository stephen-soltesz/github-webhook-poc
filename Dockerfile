FROM golang:1.11 as build
COPY . /go/src/github.com/stephen-soltesz/github-webhook-poc/
RUN go get -v github.com/stephen-soltesz/github-webhook-poc/cmd/github_webhook_receiver

# Now copy the built image into the minimal base image
#FROM alpine
#COPY --from=build /go/bin/github_webhook_receiver /
RUN cp /go/bin/github_webhook_receiver /
WORKDIR /
ENTRYPOINT ["/github_webhook_receiver"]
