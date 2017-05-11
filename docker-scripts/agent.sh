#!/bin/sh

/opt/bin/tools/sleep 10
/opt/bin/sensu-agent start --backend-url ws://backend1:8081 --subscriptions test
