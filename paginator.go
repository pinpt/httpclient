package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strings"
)

type noPaginator struct {
}

// make sure it implements the interface
var _ Paginator = (*noPaginator)(nil)

func (noPaginator) HasMore(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
	return false, nil
}

// NoPaginator returns a no-op paginator that doesn't do pagination
func NoPaginator() Paginator {
	return &noPaginator{}
}

type linkPaginator struct {
}

// make sure it implements the interface
var _ Paginator = (*linkPaginator)(nil)

var re = regexp.MustCompile("<(.*)>")

func (linkPaginator) HasMore(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
	link := resp.Header.Get("link")
	if link != "" {
		for _, token := range strings.Split(link, ", ") {
			if strings.Contains(token, "rel=\"next\"") {
				url := re.FindStringSubmatch(token)
				if len(url) > 1 {
					newreq, _ := http.NewRequest(req.Method, url[1], nil)
					newreq.Header = req.Header
					return true, newreq
				}
			}
		}
	}
	return false, nil
}

// NewLinkPaginator returns a new Paginator that uses the HTTP Link Header for building the next request
func NewLinkPaginator() Paginator {
	return &linkPaginator{}
}

type inBodyPaginator struct {
}

// make sure it implements the interface
var _ Paginator = (*inBodyPaginator)(nil)

func (inBodyPaginator) HasMore(page int, req *http.Request, resp *http.Response) (bool, *http.Request) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, nil
	}
	var mapBody map[string]interface{}
	if err := json.Unmarshal(body, &mapBody); err != nil {
		return false, nil
	}
	if v, exists := mapBody["paging"]; exists {
		var P struct {
			PageIndex int     `json:"pageIndex"`
			PageSize  float64 `json:"pageSize"`
			Total     float64 `json:"total"`
		}
		str := strings.Replace(fmt.Sprintf("%#v", v), "map[string]interface {}", "", 1)
		if err := json.Unmarshal([]byte(str), &P); err != nil || P.PageSize <= 0 || P.Total <= 0 {
			return false, nil
		}

		delete(mapBody, "paging")
		if newBody, err := json.Marshal(mapBody); err != nil {
			return false, nil
		} else {
			newBody = append(newBody, ',')
			resp.Body = ioutil.NopCloser(bytes.NewReader(newBody))
		}

		if totalPages := int(math.Ceil(P.Total / P.PageSize)); P.PageIndex < totalPages {
			index := fmt.Sprintf("p=%v", P.PageIndex)
			newIndex := fmt.Sprintf("p=%v", (P.PageIndex + 1))
			newURL := strings.Replace(req.URL.String(), index, newIndex, 1)
			newreq, _ := http.NewRequest(req.Method, newURL, nil)
			if user, pass, ok := req.BasicAuth(); ok {
				newreq.SetBasicAuth(user, pass)
			}
			return true, newreq
		}
	} else {
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	return false, nil
}

// InBodyPaginator returns a new paginator that it's pagination info it's inside body
func InBodyPaginator() Paginator {
	return &inBodyPaginator{}
}
