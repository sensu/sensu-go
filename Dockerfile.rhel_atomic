FROM registry.access.redhat.com/rhel-atomic
MAINTAINER Sensu, Inc. Engineering <engineering@sensu.io>

LABEL name="sensu/sensu-rhel" \
      maintainer="engineering@sensu.io" \
      vendor="Sensu, Inc." \
      version="2.0" \
      release="1" \
      summary="Sensu 2.0 - Full-stack monitoring" \
      description="Sensu is an event pipeline and monitoring system for everything from the server closet to the serverless application." \
      url="https://sensu.io/" \
      run="docker run -d --name sensu-backend sensu/sensu-rhel" \
      io.k8s.description="Sensu" \
      io.k8s.display-name="Sensu" \
      io.openshift.expose-services="8081:http,8080:http,3000:http,2379:http" \
      io.openshift.tags="sensu,monitoring,observability"


VOLUME /var/lib/sensu

EXPOSE 2379 2380 8080 8081 3000

RUN curl -o /usr/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 && \
    chmod +x /usr/bin/dumb-init && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-agent && \
    ln -sf /opt/sensu/bin/sensu-entrypoint.sh /usr/local/bin/sensu-backend && \
    ln -sf /opt/sensu/bin/sensuctl /usr/local/bin/sensuctl

COPY licenses /licenses
COPY target/linux-amd64/ /opt/sensu/bin/
COPY docker-scripts/ /opt/sensu/bin/

CMD ["sensu-backend"]
