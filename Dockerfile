FROM golang:1.13.7-alpine3.11 AS build

RUN apk add --no-cache git

WORKDIR $GOPATH/src/github.com/mchirico/gsa

# Copy the entire project and build it

COPY . $GOPATH/src/github.com/mchirico/gsa
RUN go get -v -t -d .
RUN go build -o /bin/project

# This results in a single layer image
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /bin/project /bin/project
ENTRYPOINT ["/bin/project"]
# Args to project
CMD ["--help"]

