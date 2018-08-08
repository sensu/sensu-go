FROM alpine:3.8 as etcdctl-fetch

RUN apk add --no-cache curl

RUN curl -fsSLO https://github.com/coreos/etcd/releases/download/v3.3.9/etcd-v3.3.9-linux-amd64.tar.gz && \
  tar xvf etcd-v3.3.9-linux-amd64.tar.gz --strip-components 1 etcd-v3.3.9-linux-amd64/etcd etcd-v3.3.9-linux-amd64/etcdctl


FROM alpine:3.8
MAINTAINER Sensu, Inc. Engineering <engineering@sensu.io>

LABEL name="sensu/sensu" \
      maintainer="engineering@sensu.io" \
      vendor="Sensu, Inc." \
      version="2.0" \
      release="1" \
      summary="Sensu 2.0 - Full-stack monitoring" \
      description="Sensu is an event pipeline and monitoring system for everything from the server closet to the serverless application." \
      url="https://sensu.io/" \
      run="docker run -d --name sensu-backend sensu/sensu" \
      io.k8s.description="Sensu" \
      io.k8s.display-name="Sensu" \
      io.openshift.expose-services="8081:http,8080:http,3000:http,2379:http" \
      io.openshift.tags="sensu,monitoring,observability"

VOLUME /var/lib/sensu

EXPOSE 2379 2380 8080 8081 3000

RUN apk add --no-cache ca-certificates dumb-init && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-agent && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-backend && \
    ln -sf /opt/sensu/bin/sensuctl /usr/local/bin/sensuctl

COPY --from=etcdctl-fetch etcdctl /usr/local/bin/etcdctl

COPY target/linux-amd64/ /opt/sensu/bin/
COPY docker-scripts/ /opt/sensu/bin/

CMD ["sensu-backend"]
