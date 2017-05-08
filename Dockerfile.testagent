FROM alpine

ENV SENSU_ID ""
ENV SENSU_BACKEND_URL ""
ENV SENSU_SUBSCRIPTIONS ""

COPY target/linux-amd64/sensu-agent /opt/sensu/bin/sensu-agent
COPY examples/checks/http_check.sh /opt/sensu/checks/check.sh

RUN apk update && apk add curl bash

EXPOSE 2379 2380 8080 8081
