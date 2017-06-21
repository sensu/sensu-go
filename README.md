# Sensu 2.0

[![Build Status](https://travis-ci.com/sensu/sensu-go.svg?token=bQ4K7jzHALx4myyBoqcu&branch=master)](https://travis-ci.com/sensu/sensu-go)

[Engineering Wiki](https://github.com/sensu/engineering/wiki)

## Sensu Agent

### Assets

#### Archive Format Specification

A valid asset archive may contain the following directories:

```
<path_to_asset>/bin # automatically prepended to the PATH environment variable
<path_to_asset>/include # automatically prepended to the CPATH environment variable
<path_to_asset>/lib # automatically prepended to the LD_LIBRARY_PATH environment variable
```

Files within these three directories will be available through the corresponding
environment variable. Any other directory will be ignored.

## Sensu Backend

### API

#### Checks

##### Create a check

```
curl -i -X POST -H 'Content-Type: application/json' -d '{"name": "check1", "interval": 60, "command": "echo 0", "subscriptions": "linux", "organization": "default"}' http://127.0.0.1:8080/checks
```

#### Events

##### Update an event

Also used to create events.

```
curl -i -X PUT -H 'Content-Type: application/json' -d '{"check": {"name": "check1", "interval": 60, "command": "echo 0"}, "entity": {"id": "scotch.local"}, "timestamp": 1493114080}' http://127.0.0.1:8080/events
```

#### Users

##### Create a user

```
curl -i -X PUT -H 'Content-Type: application/json' -d '{"username": "foo", "password": "P@ssw0rd!"}' http://127.0.0.1:8080/users
```

### Configuration

#### API Authentication

Use the `--api-authentication` flag with the **sensu-backend** binary.

## Sensu CLI
