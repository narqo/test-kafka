FROM golang:1.13-alpine AS gobuild
RUN apk update \
    && apk add --no-cache binutils musl-dev gcc
WORKDIR /go/src/test-kafka
COPY . .
RUN go build -mod=vendor ./

FROM alpine
COPY --from=gobuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=gobuild /go/src/test-kafka/test-kafka /usr/local/bin/
CMD test-kafka
