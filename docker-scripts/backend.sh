#!/bin/sh

/opt/bin/sensu-backend start --store-client-url http://0.0.0.0:2379 &
/opt/bin/tools/sleep 5
/opt/bin/sensu-cli handler create slack -t pipe --baseURL http://backend1:8080 -c "/opt/bin/handlers/slack -w https://hooks.slack.com/services/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 -c '#sensu-spam'"
/opt/bin/sensu-cli check create false -c /opt/bin/tools/false -s test --handlers slack --baseURL http://backend1:8080 -i 10

# wait for sensu-backend pid to finish so we don't kill the container
# prematurely
pid=`pgrep sensu-backend`
while kill -0 "$pid"; do
    sleep 1
done
