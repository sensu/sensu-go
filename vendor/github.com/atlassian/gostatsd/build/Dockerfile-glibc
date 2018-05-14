FROM ubuntu:16.04

RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install ca-certificates -y && \
    rm -rf /var/lib/apt/lists/*

ADD bin/linux/gostatsd /bin/gostatsd

ENTRYPOINT ["gostatsd"]
