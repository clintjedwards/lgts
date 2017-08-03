FROM golang:onbuild AS build-env
ADD . /src
RUN cd /src && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o lgts

FROM alpine
WORKDIR /app
COPY --from=build-env /src/lgts /app/
ENTRYPOINT ["./lgts"]

