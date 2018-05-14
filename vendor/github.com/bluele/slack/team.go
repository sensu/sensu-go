package slack

import (
	"encoding/json"
	"errors"
)

type TeamInfo struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"email_domain"`
	Icon        *Icon  `json:"icon"`
}

type Icon struct {
	Image34      string `json:"image_34"`
	Image44      string `json:"image_44"`
	Image68      string `json:"image_68"`
	Image88      string `json:"image_88"`
	Image102     string `json:"image_102"`
	Image132     string `json:"image_132"`
	ImageDefault bool   `json:"image_default"`
}

type TeamInfoResponse struct {
	BaseAPIResponse
	*TeamInfo `json:"team"`
}

func (sl *Slack) TeamInfo() (*TeamInfo, error) {
	uv := sl.urlValues()
	body, err := sl.GetRequest(teamInfoApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(TeamInfoResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res.TeamInfo, nil
}
