package ctxclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type (
	Context struct {
		Name        string          `json:"name,omitempty"`
		Notes       json.RawMessage `json:"notes,omitempty"`
		Pk          string          `json:"pk,omitempty"`
		Sk          string          `json:"sk,omitempty"`
		ParentId    string          `json:"parentId,omitempty"`
		LastContext string          `json:"lastContext,omitempty"`
		NoteString  string          `json:"noteString,omitempty"`
		Created     string          `json:"created,omitempty"`
		Completed   string          `json:"completed,omitempty"`
		Random      string          `json:"random,omitempty"`
	}
	ContextClient struct {
		baseUrl string
	}
	FormattedContext struct {
		Name        string             `json:"name,omitempty"`
		Notes       json.RawMessage    `json:"notes,omitempty"`
		Pk          string             `json:"pk,omitempty"`
		Sk          string             `json:"sk,omitempty"`
		SubContexts []FormattedContext `json:"subContexts,omitempty"`
		Created     string             `json:"created,omitempty"`
		Completed   string             `json:"completed,omitempty"`
		TimeSpent   TimeSpent          `json:"timeSpent,omitempty"`
	}
	TimeSpent struct {
		Time int    `json:"time,omitempty"`
		Unit string `json:"unit,omitempty"`
	}
)

var (
	SkDateFormat = "2006-01-02T15:04:05Z"
)

func NewContextClient(host, user string) *ContextClient {
	return &ContextClient{
		baseUrl: fmt.Sprintf("%s/context/%s", host, user),
	}
}

func (ctxClient *ContextClient) GetContext(contextId string) (*Context, error) {
	url := ctxClient.baseUrl
	if strings.Contains(contextId, "#") {
		ctxTimestamp := strings.Split(contextId, "#")
		timestamp := ctxTimestamp[1]
		url = fmt.Sprintf("%s?timestamp=%s", url, timestamp)
	} else if contextId != "context" {
		url = fmt.Sprintf("%s?timestamp=%s", url, contextId)
	}
	c := Context{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Handle error
		return nil, err
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	// fmt.Printf("GetCurrentContext resp\n%+v\n----\n", resp)
	if err != nil {
		// Handle error
		return nil, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	// fmt.Printf("GetCurrentContext body\n%+v\n----\n", string(body))
	if err != nil {
		// Handle error
		return nil, err
	}

	// Check the response status code
	if resp.StatusCode != 200 {
		// Handle error
		if resp.StatusCode == 404 {
			return nil, errors.Errorf("context with id '%s' not found", contextId)
		}
		return nil, errors.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	// Do something with the response body
	// fmt.Println(string(body))
	err = json.Unmarshal(body, &c)
	if err != nil {
		fmt.Printf("error unmarshaling json: %s\n", err)
		// Handle error
		return nil, err
	}
	return &c, nil
}

func (ctxClient *ContextClient) ListContexts(filterParams QSParams) (*[]Context, error) {
	url := ctxClient.baseUrl + "/list"
	url = addQSParams(url, filterParams)
	// if strings.Contains(contextId, "#") {
	// 	ctxTimestamp := strings.Split(contextId, "#")
	// 	timestamp := ctxTimestamp[1]
	// 	url = fmt.Sprintf("%s?timestamp=%s", url, timestamp)
	// } else if contextId != "context" {
	// 	url = fmt.Sprintf("%s?timestamp=%s", url, contextId)
	// }
	c := []Context{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// Handle error
		return nil, err
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	// fmt.Printf("GetCurrentContext resp\n%+v\n----\n", resp)
	if err != nil {
		// Handle error
		return nil, err
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	// fmt.Printf("GetCurrentContext body\n%+v\n----\n", string(body))
	if err != nil {
		// Handle error
		return nil, err
	}

	// Check the response status code
	if resp.StatusCode != 200 {
		// Handle error
		return nil, errors.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	// Do something with the response body
	// fmt.Println(string(body))
	err = json.Unmarshal(body, &c)
	if err != nil {
		fmt.Printf("error unmarshaling json: %s\n", err)
		// Handle error
		return nil, err
	}
	sort.Slice(c, func(i, j int) bool {
		return c[i].Created < c[j].Created
	})
	return &c, nil
}

func (ctxClient *ContextClient) GetCurrentContext() (*Context, error) {
	return ctxClient.GetContext("context")
}

func (ctxClient *ContextClient) GetLastContext() (*Context, error) {
	return ctxClient.GetContext("lastContext")
}

func (ctxClient *ContextClient) UpdateContext(c *Context) (string, error) {
	url := ctxClient.baseUrl
	// fmt.Printf("hitting url: %s\n", url)
	// fmt.Printf("url\n%s\n", url)
	bodyBytes, err := json.Marshal(c)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error marshaling request body: %s", err.Error()))
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error sending request: %s", err.Error()))
	}
	if resp.StatusCode != 200 {
		// Handle error
		return "", errors.New(fmt.Sprintf("non 200 status code: %d", resp.StatusCode))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error reading response body: %s", err.Error()))
	}
	// fmt.Printf("UpdateContext resp\n%+v\n----\n", string(respBody))
	responseContext := Context{}
	err = json.Unmarshal(respBody, &responseContext)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error unmarshaling response body: %s", err.Error()))
	}
	return responseContext.Sk, nil
	// return "new contextId", nil
}

func (ctxClient *ContextClient) CloseContext(contextId string) (string, error) {
	url := ctxClient.baseUrl + "/" + contextId + "/close"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error creating request: %s", err.Error()))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error sending request: %s", err.Error()))
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error reading response body: %s", err.Error()))
	}
	if resp.StatusCode != 200 {
		// Handle error
		return string(respBody), errors.New(fmt.Sprintf("non 200 status code: %d", resp.StatusCode))
	}
	return string(respBody), nil
	// return "new contextId", nil
}

// func (ctxClient *ContextClient) ListFormattedContexts(filterParams QSParams) (map[string]FormattedContext, error) {
// 	formatted := map[string]FormattedContext{}
// 	cs, err := ctxClient.ListContexts(filterParams)
// 	cList := *cs
// 	if err != nil {
// 		return nil, err
// 	}

// 	sort.Slice(cList, func(i, j int) bool {
// 		return cList[i].Created > cList[j].Created
// 	})

// 	// for _, c := range cs {

// 	// }

// 	return formatted, nil
// }

// func formatContext(c *Context) FormattedContext {
// 	formatted := FormattedContext{}
// 	return formatted
// }
