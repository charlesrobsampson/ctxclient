package ctxclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type (
	Queue struct {
		Name       string          `json:"name,omitempty"`
		Notes      json.RawMessage `json:"notes,omitempty"`
		Id         string          `json:"id,omitempty"`
		UserId     string          `json:"userId,omitempty"`
		NoteString string          `json:"noteString,omitempty"`
		Created    string          `json:"created,omitempty"`
		Started    string          `json:"started,omitempty"`
		ContextId  string          `json:"contextId,omitempty"`
	}

	QueueClient struct {
		baseUrl string
	}
)

var (
	QIdString = "queueId"
)

func NewQueueClient(host, user string) *QueueClient {
	return &QueueClient{
		baseUrl: fmt.Sprintf("%s/queue/%s", host, user),
	}
}

func (qClient *QueueClient) GetQueue(queueId string) (*Queue, error) {
	url := qClient.baseUrl + "/" + queueId
	q := Queue{}
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
			return nil, errors.Errorf("queue with id '%s' not found", queueId)
		}
		return nil, errors.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	// Do something with the response body
	// fmt.Println(string(body))
	err = json.Unmarshal(body, &q)
	if err != nil {
		fmt.Printf("error unmarshaling json: %s\n", err)
		// Handle error
		return nil, err
	}
	return &q, nil
}

func (qClient *QueueClient) ListQueue() (*[]Queue, error) {
	url := qClient.baseUrl
	q := []Queue{}
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
		// if resp.StatusCode == 404 {
		// 	return nil, errors.Errorf("queue with id '%s' not found", queueId)
		// }
		return nil, errors.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	// Do something with the response body
	// fmt.Println(string(body))
	err = json.Unmarshal(body, &q)
	if err != nil {
		fmt.Printf("error unmarshaling json: %s\n", err)
		// Handle error
		return nil, err
	}
	return &q, nil
}

func (qClient *QueueClient) UpdateQueue(c *Queue) (string, error) {
	url := qClient.baseUrl
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
	responseContext := Queue{}
	err = json.Unmarshal(respBody, &responseContext)
	if err != nil {
		// Handle error
		return "", errors.New(fmt.Sprintf("error unmarshaling response body: %s", err.Error()))
	}
	return responseContext.Id, nil
	// return "new contextId", nil
}

func (qClient *QueueClient) StartQueue(queueId, contextId string) (*Queue, error) {
	url := qClient.baseUrl + "/" + queueId + "/start"
	if contextId != "" {
		url += "?contextId=" + contextId
	}
	q := Queue{}
	req, err := http.NewRequest("POST", url, nil)
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
			return nil, errors.Errorf("queue with id '%s' not found", queueId)
		}
		return nil, errors.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	// Do something with the response body
	// fmt.Println(string(body))
	err = json.Unmarshal(body, &q)
	if err != nil {
		fmt.Printf("error unmarshaling json: %s\n", err)
		// Handle error
		return nil, err
	}
	return &q, nil
}
