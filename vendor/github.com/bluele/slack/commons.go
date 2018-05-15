package slack

type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int    `json:"last_set"`
}

type Purpose struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int    `json:"last_set"`
}

type BaseAPIResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}
