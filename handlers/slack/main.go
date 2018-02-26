package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bluele/slack"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var (
	webhookURL string
	channel    string
	username   string
	iconURL    string
	timeout    int
	stdin      *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err.Error())
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handler-slack",
		Short: "a slack handler built for use with sensu",
		RunE:  run,
	}

	cmd.Flags().StringVarP(&webhookURL,
		"webhook-url",
		"w",
		"",
		"The webhook url to send messages to")

	cmd.Flags().StringVarP(&channel,
		"channel",
		"c",
		"#general",
		"The channel to post messages to")

	cmd.Flags().StringVarP(&username,
		"username",
		"u",
		"sensu",
		"The username that messages will be sent as")

	cmd.Flags().StringVarP(&iconURL,
		"icon-url",
		"i",
		"http://s3-us-west-2.amazonaws.com/sensuapp.org/sensu.png",
		"A URL to an image to use as the user avatar")

	cmd.Flags().IntVarP(&timeout,
		"timeout",
		"t",
		10,
		"The amount of seconds to wait before terminating the handler")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}
	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err.Error())
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", eventJSON)
	}

	if err = validateEvent(event); err != nil {
		return errors.New(err.Error())
	}

	if err = sendMessage(event); err != nil {
		return errors.New(err.Error())
	}

	return nil
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
	hook := slack.NewWebHook(webhookURL)
	return hook.PostMessage(&slack.WebHookPostPayload{
		Text:        "",
		Channel:     channel,
		Attachments: []*slack.Attachment{messageAttachment(event)},
	})
}

func validateEvent(event *types.Event) error {
	if event.Timestamp <= 0 {
		return errors.New("timestamp is missing or must be greater than zero")
	}

	if event.Entity == nil {
		return errors.New("entity is missing from event")
	}

	if event.Check == nil {
		return errors.New("check is missing from event")
	}

	if err := event.Entity.Validate(); err != nil {
		return err
	}

	if err := event.Check.Validate(); err != nil {
		return errors.New(err.Error())
	}

	return nil
}
