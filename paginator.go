package httpclient

import (
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
