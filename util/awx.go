package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tliron/go-ard"
	"github.com/tliron/go-transcribe"
)

// See: https://ansible.readthedocs.io/projects/awx/en/latest/rest_api/api_ref.html

//
// AwxClient
//

type AwxClient struct {
	Host     string
	Username string
	Password string

	client http.Client
}

func NewAwxClient(host string, username string, password string) *AwxClient {
	if strings.HasSuffix(host, "/") {
		host = host[:len(host)-1]
	}

	self := AwxClient{
		Host:     host,
		Username: username,
		Password: password,
	}

	// See: https://stackoverflow.com/a/31309385
	self.client = http.Client{
		CheckRedirect: self.checkRedirect,
	}

	return &self
}

// Inventories

func (self *AwxClient) ListInventories(query string) (ard.Value, error) {
	path := "inventories/?" + query
	return self.Get(path)
}

func (self *AwxClient) InventoryIdFromName(name string) (int64, bool, error) {
	if results, err := self.ListInventories("name=" + name); err == nil {
		id, ok := getResultsId(results)
		return id, ok, nil
	} else {
		return 0, false, err
	}
}

// Job Templates

func (self *AwxClient) ListJobTemplates(query string) (ard.Value, error) {
	path := "job_templates/?" + query
	return self.Get(path)
}

func (self *AwxClient) JobTemplateIdFromName(name string) (int64, bool, error) {
	if results, err := self.ListJobTemplates("name=" + name); err == nil {
		id, ok := getResultsId(results)
		return id, ok, nil
	} else {
		return 0, false, err
	}
}

func (self *AwxClient) LaunchJobTemplate(id int64, inventoryId int64, extraVars map[string]any) (ard.Value, error) {
	path := fmt.Sprintf("job_templates/%d/launch/", id)
	body := make(ard.StringMap)
	if inventoryId >= 0 {
		body["inventory"] = inventoryId
	}
	if extraVars != nil {
		body["extra_vars"] = extraVars
	}
	return self.Post(path, body)
}

// Raw Requests

func (self *AwxClient) Get(path string) (ard.Value, error) {
	if request, err := self.newRequest("GET", path, nil); err == nil {
		if response, err := self.client.Do(request); err == nil {
			return ard.ReadJSON(response.Body, true)
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *AwxClient) Post(path string, body any) (ard.Value, error) {
	if body, err := toJsonReader(body); err == nil {
		if request, err := self.newRequest("POST", path, body); err == nil {
			if response, err := self.client.Do(request); err == nil {
				if response.StatusCode != 201 {
					return nil, NewAwxError(response)
				}
				return ard.ReadJSON(response.Body, true)
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (self *AwxClient) newRequest(method string, path string, body io.Reader) (*http.Request, error) {
	if request, err := http.NewRequest(method, self.Host+"/api/v2/"+path, body); err == nil {
		self.initializeRequest(request)
		return request, nil
	} else {
		return nil, err
	}
}

func (self *AwxClient) initializeRequest(request *http.Request) {
	request.SetBasicAuth(self.Username, self.Password)
	request.Header.Set("Accept", "application/json")
	if request.Body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
}

func (self *AwxClient) checkRedirect(request *http.Request, via []*http.Request) error {
	self.initializeRequest(request)
	return nil
}

//
// AwxError
//

type AwxError struct {
	StatusCode int
	Body       ard.Value
}

func NewAwxError(response *http.Response) *AwxError {
	self := AwxError{StatusCode: response.StatusCode}
	if body, err := ard.ReadJSON(response.Body, true); err == nil {
		self.Body = body
	}
	return &self
}

func (self *AwxError) Error() string {
	if self.Body != nil {
		if body, err := transcribe.NewTranscriber().Stringify(self.Body); err == nil {
			body = strings.TrimRight(body, "\n")
			return fmt.Sprintf("%d\n%s", self.StatusCode, body)
		}
	}

	return fmt.Sprintf("%d", self.StatusCode)
}

// Utils

func toJsonReader(content any) (io.Reader, error) {
	if content == nil {
		return nil, nil
	}

	if content, err := json.Marshal(content); err == nil {
		return bytes.NewReader(content), nil
	} else {
		return nil, err
	}
}

func getResultsId(results ard.Value) (int64, bool) {
	if results, ok := ard.With(results).Get("results").List(); ok {
		if len(results) == 1 {
			if id, ok := ard.With(results[0]).Get("id").ConvertSimilar().Integer(); ok {
				return id, true
			} else {
				return 0, false
			}
		} else {
			return 0, false
		}
	} else {
		return 0, false
	}
}
