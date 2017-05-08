package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bluele/slack"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

/*type SlackHandler struct {
	webhookUrl string
	channel    string
	username   string
	iconUrl    string
	timeout    int
	event      *types.Event
}*/

var (
	webhookUrl string
	channel    string
	username   string
	iconUrl    string
	timeout    int

	rootCmd = &cobra.Command{
		Use:   "handler-slack",
		Short: "a slack handler built for use with sensu",
	}
)

func init() {
	rootCmd.Flags().StringVarP(&webhookUrl,
		"webhook-url",
		"w",
		"",
		"The webhook url to send messages to")

	rootCmd.Flags().StringVarP(&channel,
		"channel",
		"c",
		"#general",
		"The channel to post messages to")

	rootCmd.Flags().StringVarP(&username,
		"username",
		"u",
		"",
		"The username that messages will be sent as")

	rootCmd.Flags().StringVarP(&iconUrl,
		"icon-url",
		"i",
		"http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png",
		"A URL to an image to use as the user avatar")

	rootCmd.Flags().IntVarP(&timeout,
		"timeout",
		"t",
		10,
		"The amount of seconds to wait before terminating the handler")
}

func formattedEventAction(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "RESOLVED"
	default:
		return "ALERT"
	}
}

func chomp(s string) string {
	return strings.Trim(strings.Trim(strings.Trim(s, "\n"), "\r"), "\r\n")
}

func eventKey(event *types.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.ID, event.Check.Name)
}

func eventSummary(event *types.Event, maxLength int) string {
	output := chomp(event.Check.Output)
	if len(event.Check.Output) > maxLength {
		output = output[0:maxLength] + "..."
	}
	return fmt.Sprintf("%s:%s", eventKey(event), output)
}

func formattedMessage(event *types.Event) string {
	return fmt.Sprintf("%s - %s", formattedEventAction(event), eventSummary(event, 100))
}

func messageColor(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "good"
	case 2:
		return "danger"
	default:
		return "warning"
	}
}

func messageStatus(event *types.Event) string {
	switch event.Check.Status {
	case 0:
		return "Resolved"
	case 2:
		return "Critical"
	default:
		return "Warning"
	}
}

func messageAttachment(event *types.Event) *slack.Attachment {
	attachment := &slack.Attachment{
		Title:    "Description",
		Text:     event.Check.Output,
		Fallback: formattedMessage(event),
		Color:    messageColor(event),
		Fields: []*slack.AttachmentField{
			&slack.AttachmentField{
				Title: "Status",
				Value: messageStatus(event),
				Short: false,
			},
			&slack.AttachmentField{
				Title: "Entity",
				Value: event.Entity.ID,
				Short: true,
			},
			&slack.AttachmentField{
				Title: "Check",
				Value: event.Check.Name,
				Short: true,
			},
		},
	}
	return attachment
}

func sendMessage(event *types.Event) error {
	hook := slack.NewWebHook(webhookUrl)
	err := hook.PostMessage(&slack.WebHookPostPayload{
		Text:        "",
		Channel:     channel,
		Attachments: []*slack.Attachment{messageAttachment(event)},
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}

	eventJson, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("failed to read stdin: ", err.Error())
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJson, event)
	if err != nil {
		log.Fatal("failed to unmarshal stdin data")
	}

	if err = sendMessage(event); err != nil {
		os.Exit(1)
	}
}
