FROM golang:1.10 AS build-env
RUN go get github.com/gorilla/mux && \
    go get github.com/gorilla/handlers && \
    go get github.com/kelseyhightower/envconfig && \
    go get github.com/nlopes/slack && \
    go get github.com/hashicorp/go-retryablehttp && \
    go get github.com/go-ozzo/ozzo-validation && \
    go get github.com/go-ozzo/ozzo-validation/is

ADD . /go/src/github.com/clintjedwards/snark
WORKDIR /go/src/github.com/clintjedwards/snark

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /src/lgts

FROM alpine
WORKDIR /app
RUN apk update && apk add ca-certificates
COPY --from=build-env /src/snark /app/
ENTRYPOINT ["./snark"]
