FROM alpine:3.6

VOLUME /var/lib/sensu

EXPOSE 2379 2380 8080 8081 3000

RUN apk add --no-cache ca-certificates dumb-init && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-agent && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-backend && \
    ln -sf /opt/sensu/bin/sensuctl /usr/local/bin/sensuctl

COPY target/linux-amd64/ /opt/sensu/bin/
COPY docker-scripts/ /opt/sensu/bin/

CMD ["sensu-backend", "start"]
