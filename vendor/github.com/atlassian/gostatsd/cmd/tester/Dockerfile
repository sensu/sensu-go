FROM alpine:3.3
RUN apk --no-cache add \
	ca-certificates
ADD statsd-tester-linux /bin/statsd-tester
ENTRYPOINT ["statsd-tester"]
