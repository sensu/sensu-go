package slack

// https://api.slack.com/docs/attachments
// It is possible to create more richly-formatted messages using Attachments.
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Color    string `json:"color,omitempty"`
	Fallback string `json:"fallback"`

	AuthorName    string `json:"author_name,omitempty"`
	AuthorSubname string `json:"author_subname,omitempty"`
	AuthorLink    string `json:"author_link,omitempty"`
	AuthorIcon    string `json:"author_icon,omitempty"`

	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Pretext   string `json:"pretext,omitempty"`
	Text      string `json:"text"`

	ImageURL string `json:"image_url,omitempty"`
	ThumbURL string `json:"thumb_url,omitempty"`

	Footer     string `json:"footer,omitempty"`
	FooterIcon string `json:"footer_icon,omitempty"`
	TimeStamp  int64  `json:"ts,omitempty"`

	Fields     []*AttachmentField `json:"fields,omitempty"`
	MarkdownIn []string           `json:"mrkdwn_in,omitempty"`
}
