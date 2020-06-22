# build stage
FROM golang:1.13-buster AS build-env
ARG APPVERSION=latest
WORKDIR /go/src/github.com/sensu/sensu-go
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/sensu/sensu-go/version.Version=$APPVERSION" -o _output/sensu-backend ./cmd/sensu-backend

FROM alpine:3.6
ENV USER=sensu
COPY --from=build-env /go/src/github.com/sensu/sensu-go/_output/sensu-backend /usr/local/bin/sensu-backend
RUN apk add --no-cache --update ca-certificates && \
    apk add wget && \
    addgroup -g 1000 ${USER} && \
    adduser -D -g "${USER} user" -H -h "/app" -G "${USER}" -u 1000 ${USER} && \
    chown -R ${USER}:${USER} /usr/local/bin/sensu-backend && \
    mkdir -p /var/lib/sensu/etcd && \
    chown -R ${USER}:${USER} /var/lib/sensu/etcd && \
    mkdir -p /var/cache/sensu && \
    chmod g+rwx /var/cache/sensu && \
    chown -R ${USER}:${USER} /var/cache/sensu

USER ${USER}:${USER}
