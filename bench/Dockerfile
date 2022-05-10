FROM golang:1.18.1-bullseye AS build

WORKDIR /src/

COPY . /src/
RUN go build -o /bin/traffic /src/cmd/traffic
RUN go build -o /bin/watch /src/cmd/watch

FROM postgres:latest

RUN apt-get update && \
    apt-get install -y iproute2;

COPY ./scripts /var/bench
COPY --from=build /bin/traffic /bin/traffic
COPY --from=build /bin/watch  /bin/watch

USER postgres
ENV POSTGRES_PASSWORD=password

CMD [ "postgres", "-c", "max_connections=1000" ]

