package client

import "encoding/json"

func (client *RestClient) delete(path string) error {
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

func (client *RestClient) get(path string, obj interface{}) error {
	res, err := client.R().SetResult(obj).Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

func (client *RestClient) list(path string, objs interface{}) error {
	res, err := client.R().Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return json.Unmarshal(res.Body(), objs)
}

func (client *RestClient) post(path string, obj interface{}) error {
	res, err := client.R().SetBody(obj).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
