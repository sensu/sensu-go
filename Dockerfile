FROM alpine

RUN apk update
RUN apk add ca-certificates

COPY target/linux-amd64/ /opt/bin/
COPY docker-scripts/ /opt/bin/

EXPOSE 2379 2380 8080 8081 3000
