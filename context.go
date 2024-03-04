package ctxclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
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

func (ctxClient *ContextClient) ListFormattedContexts(filterParams QSParams) (map[string]FormattedContext, error) {
	// formatted := map[string]FormattedContext{}
	// listed := map[string]Context{}
	cs, err := ctxClient.ListContexts(filterParams)
	ordered := *cs
	if err != nil {
		return nil, err
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Created > ordered[j].Created
	})

	fmt.Println("ordered")
	// for _, c := range ordered {
	// 	fmt.Printf("created: %s\n", c.Created)
	// 	listed[c.ContextId] = c
	// }

	formatted := formatContexts(ordered)

	keys := make([]string, 0, len(formatted))
	for k := range formatted {
		keys = append(keys, k)
	}
	sorted := []FormattedContext{}

	for _, key := range keys {
		sorted = append(sorted, formatted[key])
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Created > sorted[j].Created
	})

	out, err := json.MarshalIndent(formatted, "", "  ")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error marshaling contexts to JSON: %s", err))
	}
	fmt.Printf("out:\n%+v\n----\n", string(out))

	return formatted, nil
}

func formatContexts(ordered []Context) map[string]FormattedContext {
	lookup := map[string]Context{}
	for _, c := range ordered {
		fmt.Printf("created: %s\n", c.Created)
		lookup[c.ContextId] = c
	}
	// search range
	// 2024-02-29T01:16:54Z
	// 2024-02-29T01:22:22Z
	parentLookup := map[string]Context{}
	parentalGroups := map[string][]FormattedContext{}
	// formatted := map[string]*FormattedContext{}
	output := map[string]FormattedContext{}
	scn := bufio.NewScanner(os.Stdin)

	for i := len(ordered) - 1; i >= 0; i-- {
		c := ordered[i]
		nameWithParentId := fmt.Sprintf("%s(%s)", c.Name, c.ParentId)
		cTime := getTimeDiff(c.Created, c.Completed)
		cNotes := []string{}
		if string(c.Notes) != "" {
			err := json.Unmarshal(c.Notes, &cNotes)
			if err != nil {
				panic(fmt.Sprintf("error in unmarshal c.Notes: %s\njson: %s\n", err.Error(), string(c.Notes)))
			}
		}
		ogid, inLookup := parentLookup[nameWithParentId]
		if inLookup {
			o := lookup[ogid.ContextId]
			createdTime := getTimeDiff(o.Created, o.Completed)
			fmt.Printf("c time spent\n%+v\n", cTime)
			fmt.Printf("o time spent\n%+v\n----\n", createdTime)
			oNotes := []string{}
			// combinedNotes := json.RawMessage{}
			if string(o.Notes) != "" {
				err := json.Unmarshal(o.Notes, &oNotes)
				if err != nil {
					panic(fmt.Sprintf("error in unmarshal combined.Notes: %s\njson: %s\n", err.Error(), string(o.Notes)))
				}
			}

			noteSlice := append(cNotes, oNotes...)
			fmt.Printf("new note slice\n%v\n----\n", noteSlice)
			combinedNotes, err := json.Marshal(noteSlice)
			if err != nil {
				panic(err)
			}

			if c.Created > o.Created {
				c.Created = o.Created
			}
			if c.Completed < o.Completed {
				c.Completed = o.Completed
			}
			fmt.Printf("notes to combine\nc\n%+v\no\n%+v\n----\n", c.Notes, o.Notes)
			c.Notes = combinedNotes
			o.Notes = combinedNotes
		}
	}

	l, _ := json.MarshalIndent(lookup, "", "  ")
	fmt.Printf("lookup: %s\n", string(l))

	fmt.Printf("parentLookup: %+v\n", parentLookup)
	pLook, _ := json.MarshalIndent(parentLookup, "", "  ")
	fmt.Println(string(pLook))

	for _, c := range ordered {
		if c.ParentId != "" {
			p := lookup[c.ParentId]
			pid := fmt.Sprintf("%s(%s)", p.Name, p.ParentId)
			parent := parentLookup[pid]
			fmt.Printf("on %s\nwith: %s\nold: %s\nnew: %s\n", c.Name, pid, c.ParentId, parent.ContextId)
			// c.ParentId = parentId
		}
		f := formatContext(&c)
		if c.ParentId != "" {
			parentalGroups[c.ParentId] = append(parentalGroups[c.ParentId], f)
		}
	}

	fmt.Println("parentalGroups:")
	for id, children := range parentalGroups {
		fmt.Println(id)
		for _, c := range children {
			fmt.Printf("    %s\n", c.Name)
		}
	}

	// TODO: resolve children and append to parents in formatted

	// for _, c := range list {
	// 	existing := formatted[c.ContextId]
	// 	existingWithName := formatted[c.Name]
	// 	fmt.Printf("\non name: %s id: %s\n", c.Name, c.ContextId)
	// 	fmt.Printf("%+v\n", c)
	// 	fmt.Printf("existing: %+v\n", existing)
	// 	fmt.Printf("existingWithName: %+v\n", existingWithName)
	// 	f := FormattedContext{}
	// 	if c.ParentId == "" {
	// 		fmt.Println("no parent")
	// 		if existing == nil && existingWithName == nil {
	// 			fmt.Println("no children")
	// 			f = formatContext(&c)
	// 			formatted[c.Name] = &f
	// 		}
	// 	} else {
	// 		p := formatted[c.ParentId]
	// 		if p == nil {
	// 			fmt.Println("has parent but it hasn't been defined yet")
	// 			p = &FormattedContext{
	// 				ContextId:   c.ParentId,
	// 				SubContexts: []FormattedContext{formatContext(&c)},
	// 			}
	// 			fmt.Printf("add parent to map uder id:\n%+v\n", p)
	// 			formatted[c.ParentId] = p
	// 		} else {
	// 			fmt.Println("has parent and it has been defined")
	// 			f := formatContext(&c)
	// 			p.SubContexts = append(p.SubContexts, f)
	// 			sort.Slice(p.SubContexts, func(i, j int) bool {
	// 				return p.SubContexts[i].Created < p.SubContexts[j].Created
	// 			})
	// 			fmt.Printf("update parent with new child:\n%+v\n", p)
	// 		}
	// 		// formatted[c.ContextId] = &f
	// 	}

	scn.Scan()
	// if existing != nil {
	// 	if existingWithName != nil {
	// 		fmt.Println("existingWithName")
	// 	} else {
	// 		fmt.Println("existing by id only")
	// 		f = formatContext(&c)
	// 		f.SubContexts = append(f.SubContexts, existing.SubContexts...)
	// 		sort.Slice(f.SubContexts, func(i, j int) bool {
	// 			return f.SubContexts[i].Created < f.SubContexts[j].Created
	// 		})
	// 		fmt.Printf("f: %+v\n", f)
	// 		delete(formatted, c.ContextId)
	// 	}
	// }
	// if c.ParentId == "" {

	// 	formatted[c.ContextId] = &f
	// } else {
	// 	p := formatted[c.ParentId]
	// 	if p == nil {
	// 		p = &FormattedContext{
	// 			ContextId:   c.ParentId,
	// 			SubContexts: []FormattedContext{f},
	// 		}
	// 	} else {
	// 		p.SubContexts = append(p.SubContexts, f)
	// 		sort.Slice(p.SubContexts, func(i, j int) bool {
	// 			return p.SubContexts[i].Created > p.SubContexts[j].Created
	// 		})
	// 	}
	// 	formatted[c.ParentId] = p
	// }
	// if existing != nil {
	// 	f.SubContexts = append(f.SubContexts, existing.SubContexts...)
	// } else {
	// 	f = formatContext(&c)
	// }
	// if c.ParentId == "" {
	// 	formatted[c.ContextId] = &f
	// } else {
	// 	p := formatted[c.ParentId]
	// 	if p == nil {
	// 		p = &FormattedContext{
	// 			ContextId:   c.ParentId,
	// 			SubContexts: []FormattedContext{f},
	// 		}
	// 	} else {
	// 		p.SubContexts = append(p.SubContexts, f)
	// 		sort.Slice(p.SubContexts, func(i, j int) bool {
	// 			return p.SubContexts[i].Created > p.SubContexts[j].Created
	// 		})
	// 	}
	// 	formatted[c.ParentId] = p
	// }
	// }

	// for id, c := range formatted {
	// 	output[id] = *c
	// }

	// for id, c := range cs {
	// 	if c.ParentId != "" {
	// 		f := formatContext(&c)
	// 		p := formatted[c.ParentId]
	// 		p.SubContexts = append(p.SubContexts, f)
	// 		sort.Slice(p.SubContexts, func(i, j int) bool {
	// 			return p.SubContexts[i].Created < p.SubContexts[j].Created
	// 		})
	// 		formatted[c.ParentId] = p
	// 	} else {
	// 		formatted[id] = formatContext(&c)
	// 	}
	// }
	return output
}

func formatContext(c *Context) FormattedContext {
	return FormattedContext{
		Name:      c.Name,
		Notes:     c.Notes,
		ContextId: c.ContextId,
		Created:   c.Created,
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
