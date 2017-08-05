FROM golang:onbuild AS build-env
ADD . /src
RUN cd /src && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o lgts

FROM alpine
WORKDIR /app
RUN apk update && apk add ca-certificates
COPY --from=build-env /src/lgts /app/
ENTRYPOINT ["./lgts"]

