package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
)

// API chat.postMessage: Sends a message to a channel.
func (sl *Slack) ChatPostMessage(channelId string, text string, opt *ChatPostMessageOpt) error {
	uv, err := sl.buildChatPostMessageUrlValues(opt)
	if err != nil {
		return err
	}
	uv.Add("channel", channelId)

	body, err := sl.PostRequest(chatPostMessageApiEndpoint, uv, sl.buildRequestBodyForm(text))
	if err != nil {
		return err
	}
	res := new(ChatPostMessageAPIResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return err
	}
	if !res.Ok {
		return errors.New(res.Error)
	}
	return nil
}

// option type for `chat.postMessage` api
type ChatPostMessageOpt struct {
	AsUser      bool
	Username    string
	Parse       string
	LinkNames   string
	Attachments []*Attachment
	UnfurlLinks string
	UnfurlMedia string
	IconUrl     string
	IconEmoji   string
}

// response type for `chat.postMessage` api
type ChatPostMessageAPIResponse struct {
	BaseAPIResponse
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
}

func (sl *Slack) buildRequestBodyForm(text string) (*bytes.Buffer) {
    return bytes.NewBuffer([]byte("text=" + text))
}

func (sl *Slack) buildChatPostMessageUrlValues(opt *ChatPostMessageOpt) (*url.Values, error) {
	uv := sl.urlValues()
	if opt == nil {
		return uv, nil
	}
	if opt.AsUser {
		uv.Add("as_user", "true")
	} else if opt.Username != "" {
		uv.Add("username", opt.Username)
	}
	if opt.Parse != "" {
		uv.Add("parse", opt.Parse)
	}
	if opt.LinkNames != "" {
		uv.Add("link_names", opt.LinkNames)
	}
	if opt.UnfurlLinks != "" {
		uv.Add("unfurl_links", opt.UnfurlLinks)
	}
	if opt.UnfurlMedia != "" {
		uv.Add("unfurl_media", opt.UnfurlMedia)
	}
	if opt.IconUrl != "" {
		uv.Add("icon_url", opt.IconUrl)
	}
	if opt.IconEmoji != "" {
		uv.Add("icon_emoji", opt.IconEmoji)
	}
	if opt.Attachments != nil && len(opt.Attachments) > 0 {
		ats, err := json.Marshal(opt.Attachments)
		if err != nil {
			return nil, err
		}
		uv.Add("attachments", string(ats))
	}
	return uv, nil
}
