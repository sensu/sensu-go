FROM alpine

RUN apk update
RUN apk add ca-certificates
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY target/linux-amd64/ /opt/bin/
COPY docker-scripts/ /opt/bin/

EXPOSE 2379 2380 8080 8081 3000
