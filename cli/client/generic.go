package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-resty/resty/v2"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

// Delete sends a DELETE request to the given path
func (client *RestClient) Delete(path string) error {
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// Get sends a GET request for an object at the given path
func (client *RestClient) Get(path string, obj interface{}) error {
	res, err := client.R().SetResult(obj).Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// List sends a GET request for all objects at the given path.
// The options parameter allows for enhancing the request with field/label
// selectors (filtering), pagination, ...
func (client *RestClient) List(path string, objs interface{}, options *ListOptions, header *http.Header) error {
	objsType := reflect.TypeOf(objs)
	if objsType.Kind() != reflect.Ptr || objsType.Elem().Kind() != reflect.Slice {
		panic("unexpected type for objs")
	}

	for {
		request := client.R()
		ApplyListOptions(request, options)

		resp, err := request.Get(path)
		if err != nil {
			return err
		}
		if header != nil {
			*header = resp.Header()
		}

		if resp.StatusCode() >= 400 {
			return UnmarshalError(resp)
		}

		body := resp.Body()
		if len(body) == 0 {
			return nil
		}

		switch objs.(type) {
		case *[]types.Wrapper, *[]*types.Wrapper:
			if err := json.Unmarshal(body, objs); err != nil {
				return err
			}
		default:
			o := reflect.ValueOf(objs).Elem()

			var slice []*types.Wrapper
			var wrapper types.Wrapper

			if err := json.Unmarshal(body, &slice); err == nil {
				// This case is for when the API returns a slice of wrapped resources,
				// but we've passed in unwrapped resources to be filled.
				for _, wrapper := range slice {
					o.Set(reflect.Append(o, reflect.ValueOf(wrapper.Value)))
				}
			} else if err := json.Unmarshal(body, &wrapper); err == nil {
				// This case is for when the API returns a single wrapped value, but we've
				// passed in the unwrapped value.
				o.Set(reflect.Append(o, reflect.ValueOf(wrapper.Value)))
			} else {
				newObjs := reflect.New(objsType.Elem())
				if len(body) > 0 && body[0] == '{' {
					// This case is for when the API returns a single unwrapped value.
					elem := reflect.New(reflect.Indirect(newObjs).Type().Elem().Elem())
					if err := json.Unmarshal(body, elem.Interface()); err != nil {
						return err
					}
					o.Set(reflect.Append(o, elem))
					return nil
				}

				// And this is the default, the common case.
				if err := json.Unmarshal(body, newObjs.Interface()); err != nil {
					return err
				}

				o.Set(reflect.AppendSlice(o, newObjs.Elem()))
			}
		}

		options.ContinueToken = resp.Header().Get(corev2.PaginationContinueHeader)
		if options.ContinueToken == "" {
			break
		}
	}

	return nil
}

// Post sends a POST request with obj as the payload to the given path
func (client *RestClient) Post(path string, obj interface{}) error {
	res, err := client.R().SetBody(obj).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// PostWithResponse sends a POST request with obj as the payload to the given path
// and additionally returns the response
func (client *RestClient) PostWithResponse(path string, obj interface{}) (*resty.Response, error) {
	res, err := client.R().SetBody(obj).Post(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	return res, nil
}

// Put sends a PUT request with obj as the payload to the given path
func (client *RestClient) Put(path string, obj interface{}) error {
	res, err := client.R().SetBody(obj).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// PutResource ...
func (client *RestClient) PutResource(r types.Wrapper) error {
	path := r.Value.URIPath()

	// Determine if we should send the wrapped resource or only the resource
	// itself
	var bytes []byte
	var err error
	if r.APIVersion == "core/v2" {
		bytes, err = json.Marshal(r.Value)
	} else {
		bytes, err = json.Marshal(r)
	}
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}
	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}
	return nil
}
