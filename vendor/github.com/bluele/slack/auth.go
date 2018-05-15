package slack

import (
	"encoding/json"
	"errors"
)

// API auth.test: Checks authentication and tells you who you are.
func (sl *Slack) AuthTest() (*AuthTestApiResponse, error) {
	uv := sl.urlValues()
	body, err := sl.GetRequest(authTestApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(AuthTestApiResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res, nil
}

// response type for `auth.test` api
type AuthTestApiResponse struct {
	BaseAPIResponse
	Url    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamId string `json:"team_id"`
	UserId string `json:"user_id"`
}
