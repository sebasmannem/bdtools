FROM golang:alpine AS bdtoolsbuilder
WORKDIR /usr/src/app

COPY . .
RUN apk update && \
    apk upgrade && \
    apk add git && \
    sh set_version.sh && \
    go mod tidy && \
    go build -o ./bin/bdtools ./cmd/bdtools

FROM alpine:latest

COPY --from=bdtoolsbuilder /usr/src/app/bin/bdtools /usr/local/bin/
COPY config/bdtools.yaml /etc/bdtools/config.yaml
ENTRYPOINT [ "/usr/local/bin/bdtools" ]
CMD [ "-c", "/etc/bdtools/config.yaml" ]
