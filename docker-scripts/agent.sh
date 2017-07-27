#!/bin/sh

/opt/sensu/bin/tools/sleep 10
/usr/local/bin/sensu-agent start --backend-url ws://backend1:8081 --subscriptions test
