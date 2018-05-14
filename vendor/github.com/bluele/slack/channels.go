package slack

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"
)

// API channels.list: Lists all channels in a Slack team.
func (sl *Slack) ChannelsList() ([]*Channel, error) {
	uv := sl.urlValues()
	body, err := sl.GetRequest(channelsListApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(ChannelsListAPIResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res.Channels()
}

// response type for `channels.list` api
type ChannelsListAPIResponse struct {
	BaseAPIResponse
	RawChannels json.RawMessage `json:"channels"`
}

// slack channel type
type Channel struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	IsChannel  bool            `json:"is_channel"`
	Created    int             `json:"created"`
	Creator    string          `json:"creator"`
	IsArchived bool            `json:"is_archived"`
	IsGeneral  bool            `json:"is_general"`
	IsMember   bool            `json:"is_member"`
	Members    []string        `json:"members"`
	RawTopic   json.RawMessage `json:"topic"`
	RawPurpose json.RawMessage `json:"purpose"`
	NumMembers int             `json:"num_members"`
}

// Channels returns a slice of channel object from a response of `channels.list` api.
func (res *ChannelsListAPIResponse) Channels() ([]*Channel, error) {
	var chs []*Channel
	err := json.Unmarshal(res.RawChannels, &chs)
	if err != nil {
		return nil, err
	}
	return chs, nil
}

func (ch *Channel) Topic() (*Topic, error) {
	tp := new(Topic)
	err := json.Unmarshal(ch.RawTopic, tp)
	if err != nil {
		return nil, err
	}
	return tp, nil
}

func (ch *Channel) Purpose() (*Purpose, error) {
	pp := new(Purpose)
	err := json.Unmarshal(ch.RawPurpose, pp)
	if err != nil {
		return nil, err
	}
	return pp, nil
}

// FindChannel returns a channel object that satisfy conditions specified.
func (sl *Slack) FindChannel(cb func(*Channel) bool) (*Channel, error) {
	channels, err := sl.ChannelsList()
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		if cb(channel) {
			return channel, nil
		}
	}
	return nil, errors.New("No such channel.")
}

// FindChannelByName returns a channel object that matches name specified.
func (sl *Slack) FindChannelByName(name string) (*Channel, error) {
	return sl.FindChannel(func(channel *Channel) bool {
		return channel.Name == name
	})
}

// API channels.join: Joins a channel, creating it if needed.
func (sl *Slack) JoinChannel(name string) error {
	uv := sl.urlValues()
	uv.Add("name", name)

	_, err := sl.GetRequest(channelsJoinApiEndpoint, uv)
	if err != nil {
		return err
	}
	return nil
}

type Message struct {
	Type    string `json:"type"`
	Ts      string `json:"ts"`
	UserId  string `json:"user"`
	Text    string `json:"text"`
	Subtype string `json:"subtype"`
}

func (msg *Message) Timestamp() *time.Time {
	seconds, _ := strconv.ParseInt(msg.Ts[0:10], 10, 64)
	microseconds, _ := strconv.ParseInt(msg.Ts[11:17], 10, 64)
	ts := time.Unix(seconds, microseconds*1e3)
	return &ts
}

// option type for `channels.history` api
type ChannelsHistoryOpt struct {
	Channel   string  `json:"channel"`
	Latest    float64 `json:"latest"`
	Oldest    float64 `json:"oldest"`
	Inclusive int     `json:"inclusive"`
	Count     int     `json:"count"`
	UnReads   int     `json:"unreads,omitempty"`
}

func (opt *ChannelsHistoryOpt) Bind(uv *url.Values) error {
	uv.Add("channel", opt.Channel)
	if opt.Latest != 0.0 {
		uv.Add("lastest", strconv.FormatFloat(opt.Latest, 'f', 6, 64))
	}
	if opt.Oldest != 0.0 {
		uv.Add("oldest", strconv.FormatFloat(opt.Oldest, 'f', 6, 64))
	}
	uv.Add("inclusive", strconv.Itoa(opt.Inclusive))
	if opt.Count != 0 {
		uv.Add("count", strconv.Itoa(opt.Count))
	}
	uv.Add("unreads", strconv.Itoa(opt.UnReads))
	return nil
}

// response type for `channels.history` api
type ChannelsHistoryResponse struct {
	BaseAPIResponse
	Latest             float64    `json:"latest"`
	Messages           []*Message `json:"messages"`
	HasMore            bool       `json:"has_more"`
	UnReadCountDisplay int        `json:"unread_count_display"`
}

// API channels.history: Fetches history of messages and events from a channel.
func (sl *Slack) ChannelsHistory(opt *ChannelsHistoryOpt) (*ChannelsHistoryResponse, error) {
	uv := sl.urlValues()
	err := opt.Bind(uv)
	if err != nil {
		return nil, err
	}
	body, err := sl.GetRequest(channelsHistoryApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(ChannelsHistoryResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res, nil
}

func (sl *Slack) ChannelsHistoryMessages(opt *ChannelsHistoryOpt) ([]*Message, error) {
	res, err := sl.ChannelsHistory(opt)
	if err != nil {
		return nil, err
	}
	return res.Messages, nil
}
