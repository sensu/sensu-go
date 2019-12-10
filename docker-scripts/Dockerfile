FROM ubuntu:18.04

#Build in this directory with sudo docker build . or docker build . (depending on your OS)
ENV GOROOT=/usr/local/go
ENV GOPATH=/opt/sensu
ENV PATH=$GOPATH:/bin:$GOROOT/bin:/opt/sensu/bin/:$PATH
#Using Go 1.13.3 for Go modules, change for your needs.
ARG GO_VER=1.13.3

WORKDIR /opt/sensu
COPY ./sensu-entrypoint.sh /opt/sensu/bin/

RUN apt update &&\
    apt -y upgrade &&\
    apt -y install git wget

#Download and install Go using the GO_VER argument.
RUN wget https://dl.google.com/go/go$GO_VER.linux-amd64.tar.gz && tar -xvf go$GO_VER.linux-amd64.tar.gz &&\
    mv go /usr/local/ &&\
    rm go$GO_VER.linux-amd64.tar.gz

#Build OSS only sensu features, enterprise features are built in the Sensu supported images:
#docker pull sensu/sensu (Alpine based container) or 
#docker pull sensu/sensu-rhel (Red Hat based container)
RUN git clone --depth=1 https://github.com/sensu/sensu-go.git &&\ 
    cd sensu-go &&\
    go build -o ../bin/sensu-agent ./cmd/sensu-agent &&\
    go build -o ../bin/sensu-backend ./cmd/sensu-backend &&\
    go build -o ../bin/sensuctl ./cmd/sensuctl &&\
    rm -rf /opt/sensu/sensu-go

#State directory lives outside container for persistent data.
VOLUME [/var/lib/sensu]
#Port 3000 for Sensu Go UI, 8080 for Sensu API, and port 8081 for sensu-agent websocket API communication.  
EXPOSE 3000 8080 8081

#Use Docker logs to see stdout of sensu-backend (ie. docker logs --details -f  `docker ps -q` )
CMD ["/opt/sensu/bin/sensu-backend","start"]
