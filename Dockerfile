FROM scratch

COPY target/linux-amd64/ /

EXPOSE 2379 2380 8080 8081 3000
