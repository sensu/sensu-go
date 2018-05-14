package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// API files.upload: Uploads or creates a file.
func (sl *Slack) FilesUpload(opt *FilesUploadOpt) error {
	req, err := sl.createFilesUploadRequest(opt)
	if err != nil {
		return err
	}
	body, err := sl.DoRequest(req)
	if err != nil {
		return err
	}
	res := new(FilesUploadAPIResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return err
	}
	if !res.Ok {
		return errors.New(res.Error)
	}
	return nil
}

// option type for `files.upload` api
type FilesUploadOpt struct {
	Content        string
	Filepath       string
	Filetype       string
	Filename       string
	Title          string
	InitialComment string
	Channels       []string
}

// response of `files.upload` api
type FilesUploadAPIResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func (sl *Slack) createFilesUploadRequest(opt *FilesUploadOpt) (*http.Request, error) {
	var body io.Reader

	uv := sl.urlValues()
	if opt == nil {
		return nil, errors.New("`opt *FilesUploadOpt` argument must be specified.")
	}
	contentType := "application/x-www-form-urlencoded"

	if opt.Filetype != "" {
		uv.Add("filetype", opt.Filetype)
	}
	if opt.Filename != "" {
		uv.Add("filename", opt.Filename)
	}
	if opt.Title != "" {
		uv.Add("title", opt.Title)
	}
	if opt.InitialComment != "" {
		uv.Add("initial_comment", opt.InitialComment)
	}
	if len(opt.Channels) != 0 {
		uv.Add("channels", strings.Join(opt.Channels, ","))
	}

	if opt.Filepath != "" {
		var err error
		body, contentType, err = createFileParam("file", opt.Filepath)
		if err != nil {
			return nil, err
		}
	} else if opt.Content != "" {
		body = strings.NewReader(url.Values{"content": []string{opt.Content}}.Encode())
	}

	req, err := http.NewRequest("POST", apiBaseUrl+filesUploadApiEndpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.URL.RawQuery = (*uv).Encode()
	return req, nil
}

func createFileParam(param, path string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	p, err := filepath.Abs(path)
	if err != nil {
		return nil, "", err
	}
	file, err := os.Open(p)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()
	part, err := writer.CreateFormFile(param, filepath.Base(path))
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
