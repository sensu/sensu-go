#!/usr/bin/dumb-init sh

SENSU=/opt/sensu

called=$(basename $0)

${SENSU}/bin/${called} $@