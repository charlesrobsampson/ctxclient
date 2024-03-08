package ctxclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type (
	Context struct {
		Name        string          `json:"name,omitempty"`
		Notes       json.RawMessage `json:"notes,omitempty"`
		UserId      string          `json:"userId,omitempty"`
		ContextId   string          `json:"contextId,omitempty"`
		ParentId    string          `json:"parentId,omitempty"`
		LastContext string          `json:"lastContext,omitempty"`
		NoteString  string          `json:"noteString,omitempty"`
		Created     string          `json:"created,omitempty"`
		Completed   string          `json:"completed,omitempty"`
		// TimeSpent   TimeSpent       `json:"timeSpent,omitempty"`
	}
	ContextClient struct {
		baseUrl string
	}
	FormattedContext struct {
		Name        string             `json:"name,omitempty"`
		Notes       json.RawMessage    `json:"notes,omitempty"`
		UserId      string             `json:"userId,omitempty"`
		ContextId   string             `json:"contextId,omitempty"`
		Created     string             `json:"created,omitempty"`
		Completed   string             `json:"completed,omitempty"`
		TimeSpent   TimeSpent          `json:"timeSpent,omitempty"`
		SubContexts []FormattedContext `json:"subContexts,omitempty"`
	}
	TimeSpent struct {
		Time float64 `json:"time,omitempty"`
		Unit string  `json:"unit,omitempty"`
	}
)

var (
	SkDateFormat    = "2006-01-02T15:04:05Z"
	PkString        = "userId"
	TimestampString = "contextId"
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
	} else if contextId != "current" {
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
	return ctxClient.GetContext("current")
}

func (ctxClient *ContextClient) GetLastContext() (*Context, error) {
	return ctxClient.GetContext("last")
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
	return responseContext.ContextId, nil
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

func (ctxClient *ContextClient) ListFormattedContexts(filterParams QSParams) ([]FormattedContext, error) {
	cs, err := ctxClient.ListContexts(filterParams)
	ordered := *cs
	if err != nil {
		return nil, err
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Created > ordered[j].Created
	})

	formatted := formatContexts(ordered)
	return formatted, nil
}

func formatContexts(ordered []Context) []FormattedContext {
	type ctxid struct {
		Name     string `json:"name,omitempty"`
		ParentId string `json:"parentId,omitempty"`
	}
	type newOrder struct {
		CID     ctxid
		Context FormattedContext
	}
	lookup := map[string]Context{}
	flattenedFormatted := map[ctxid][]FormattedContext{}
	consolidatedFormatted := map[ctxid]FormattedContext{}
	for _, c := range ordered {
		lookup[c.ContextId] = c
		cID := ctxid{
			Name:     c.Name,
			ParentId: c.ParentId,
		}
		existingFormatted := flattenedFormatted[cID]
		f := formatContext(&c)
		existingFormatted = append(existingFormatted, f)
		sort.Slice(existingFormatted, func(i, j int) bool {
			return existingFormatted[i].Created < existingFormatted[j].Created
		})
		flattenedFormatted[cID] = existingFormatted
	}

	theNewOrder := []newOrder{}
	for cID, l := range flattenedFormatted {
		f := FormattedContext{}
		for _, c := range l {
			c.TimeSpent = getTimeDiff(c.Created, c.Completed)
			if f.Name != "" {
				// add times together
				if f.TimeSpent.Unit == c.TimeSpent.Unit {
					f.TimeSpent.Time += c.TimeSpent.Time
				} else {
					// balance units and then add
					fmt.Printf("unbalanced units with self: %s and %s\n", f.TimeSpent.Unit, c.TimeSpent.Unit)
				}
				f.Completed = c.Completed
				fNotes := []string{}
				cNotes := []string{}
				if f.Notes != nil {
					err := json.Unmarshal(f.Notes, &fNotes)
					if err != nil {
						panic("error in unmarshal f.Notes: " + err.Error())
					}
				}
				if c.Notes != nil {
					err := json.Unmarshal(c.Notes, &cNotes)
					if err != nil {
						panic("error in unmarshal c.Notes: " + err.Error())
					}
				}
				combinedNotes := append(fNotes, cNotes...)
				if len(combinedNotes) > 0 {
					noteBytes, err := json.Marshal(combinedNotes)
					if err != nil {
						panic("error in marshal combinedNotes: " + err.Error())
					}
					f.Notes = noteBytes
				}
			} else {
				f = c
			}
		}
		consolidatedFormatted[cID] = f
		theNewOrder = append(theNewOrder, newOrder{
			CID:     cID,
			Context: f,
		})
	}

	sort.Slice(theNewOrder, func(i, j int) bool {
		return theNewOrder[i].Context.Created > theNewOrder[j].Context.Created
	})

	formatted := []FormattedContext{}
	for _, ctx := range theNewOrder {
		cID := ctx.CID
		c := consolidatedFormatted[cID]
		p := lookup[cID.ParentId]
		if p.Name != "" {
			pID := ctxid{
				Name:     p.Name,
				ParentId: p.ParentId,
			}
			parent := consolidatedFormatted[pID]

			kids := parent.SubContexts
			kids = append(kids, c)
			sort.Slice(kids, func(i, j int) bool {
				return kids[i].Created < kids[j].Created
			})
			if parent.TimeSpent.Unit == c.TimeSpent.Unit {
				parent.TimeSpent.Time += c.TimeSpent.Time
				parent.TimeSpent.Time = math.Round(parent.TimeSpent.Time*100) / 100
			} else {
				// balance units and then add
				fmt.Printf("unbalanced units with parent: %s and %s\n", parent.TimeSpent.Unit, c.TimeSpent.Unit)
			}
			parent.SubContexts = kids
			consolidatedFormatted[pID] = parent
		} else {
			formatted = append(formatted, c)
		}
	}
	sort.Slice(formatted, func(i, j int) bool {
		return formatted[i].Created < formatted[j].Created
	})
	return formatted
}

func formatContext(c *Context) FormattedContext {
	return FormattedContext{
		Name:      c.Name,
		Notes:     c.Notes,
		ContextId: c.ContextId,
		Created:   c.Created,
		Completed: c.Completed,
	}
}

func getTimeDiff(start, end string) TimeSpent {
	// Parse the time strings into time.Time objects.
	t1, err := time.Parse(SkDateFormat, start)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	time2 := end
	if time2 == "" {
		time2 = time.Now().UTC().Format(SkDateFormat)
	}
	t2, err := time.Parse(SkDateFormat, time2)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Calculate the time difference in minutes.
	diff := t2.Sub(t1).Minutes()

	// if diff < 60 {
	// 	diff *= 60
	// 	return TimeSpent{
	// 		Time: math.Round(f*100) / 100,
	// 		Unit: "second",
	// 	}
	// }

	return TimeSpent{
		Time: math.Round(diff*100) / 100,
		Unit: "minute",
	}
}
