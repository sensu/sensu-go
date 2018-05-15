# Slack [![GoDoc](https://godoc.org/github.com/bluele/slack?status.png)](https://godoc.org/github.com/bluele/slack)

Golang client for the Slack API. Include the example code using each slack api.

## Currently supports:

Method | Description | Example
--- | --- | ---
channels.history | Fetches history of messages and events from a channel. | [#link](https://github.com/bluele/slack/blob/master/examples/channels_history.go)
channels.join | Joins a channel, creating it if needed. | [#link](https://github.com/bluele/slack/blob/master/examples/channels_join.go)
channels.list | Lists all channels in a Slack team. | [#link](https://github.com/bluele/slack/blob/master/examples/channels_list.go)
chat.postMessage | Sends a message to a channel. | [#link](https://github.com/bluele/slack/blob/master/examples/chat_post_message.go)
files.upload | Upload an image/file | [#link](https://github.com/bluele/slack/blob/master/examples/upload_file.go)
groups.invite | Invites a user to a private group. | [#link](https://github.com/bluele/slack/blob/master/examples/groups_invite.go)
groups.create | Creates a private group. | [#link](https://github.com/bluele/slack/blob/master/examples/groups_create.go)
groups.list | Lists private groups that the calling user has access to. | [#link](https://github.com/bluele/slack/blob/master/examples/groups_list.go)
users.info | Gets information about a channel. | [#link](https://github.com/bluele/slack/blob/master/examples/users_info.go)
users.list | Lists all users in a Slack team. | [#link](https://github.com/bluele/slack/blob/master/examples/users_list.go)


## Example

```go
package main

import (
  "github.com/bluele/slack"
)

const (
  token       = "your-api-token"
  channelName = "general"
)

func main() {
  api := slack.New(token)
  err := api.ChatPostMessage(channelName, "Hello, world!", nil)
  if err != nil {
    panic(err)
  }
}
```

## Command line tool

If you are looking for slack commandline utility, [vektorlab/slackcat](https://github.com/vektorlab/slackcat) probably suits you.

# Author

**Jun Kimura**

* <http://github.com/bluele>
* <junkxdev@gmail.com>
