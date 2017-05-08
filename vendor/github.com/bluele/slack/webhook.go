package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type WebHook struct {
	hookURL string
}

type WebHookPostPayload struct {
	Text        string        `json:"text,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	IconUrl     string        `json:"icon_url,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	UnfurlLinks bool          `json:"unfurl_links,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

func NewWebHook(hookURL string) *WebHook {
	return &WebHook{hookURL}
}

func (hk *WebHook) PostMessage(payload *WebHookPostPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(hk.hookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(t))
	}

	return nil
}
