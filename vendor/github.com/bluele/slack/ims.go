package slack

import (
	"encoding/json"
	"errors"
)

// API im.list: Lists direct message channels for the calling user.
func (sl *Slack) ImList() ([]*Im, error) {
	uv := sl.urlValues()
	body, err := sl.GetRequest(imListApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(ImListAPIResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res.Ims()
}

// slack im type
type Im struct {
	Id            string `json:"id"`
	Isim          bool   `json:"is_im"`
	User          string `json:"user"`
	Created       int    `json:"created"`
	IsUserDeleted bool   `json:"is_user_deleted"`
}

// response type for `im.list` api
type ImListAPIResponse struct {
	BaseAPIResponse
	RawIms json.RawMessage `json:"ims"`
}

// Ims returns a slice of im object from `im.list` api.
func (res *ImListAPIResponse) Ims() ([]*Im, error) {
	var im []*Im
	err := json.Unmarshal(res.RawIms, &im)
	if err != nil {
		return nil, err
	}
	return im, nil
}

// FindIm returns a im object that satisfy conditions specified.
func (sl *Slack) FindIm(cb func(*Im) bool) (*Im, error) {
	ims, err := sl.ImList()
	if err != nil {
		return nil, err
	}
	for _, im := range ims {
		if cb(im) {
			return im, nil
		}
	}
	return nil, errors.New("No such im.")
}

// FindImByName returns a im object that matches name specified.
func (sl *Slack) FindImByName(name string) (*Im, error) {
	user, err := sl.FindUserByName(name)
	if err != nil {
		return nil, err
	}
	id := user.Id
	return sl.FindIm(func(im *Im) bool {
		return im.User == id
	})
}
