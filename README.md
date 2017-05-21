# Sensu 2.0

[![Build Status](https://travis-ci.com/sensu/sensu-go.svg?token=bQ4K7jzHALx4myyBoqcu&branch=master)](https://travis-ci.com/sensu/sensu-go)

[Engineering Wiki](https://github.com/sensu/engineering/wiki)

## API

### Events

#### Update an event

Also used to create events.

```
curl -i -X PUT -H 'Content-Type: application/json' -d '{"check": {"name": "check1", "interval": 60, "command": "echo 0"}, "entity": {"id": "scotch.local"}, "timestamp": 1493114080}' http://127.0.0.1:8080/events
```

### Users

#### Create a user

```
curl -i -X PUT -H 'Content-Type: application/json' -d '{"username": "foo", "password": "P@ssw0rd!"}' http://127.0.0.1:8080/users
```

## Backend

### Configuration

#### API Authentication

Use the `--api-authentication` flag with **sensu-backend**.
