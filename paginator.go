package httpclient

import "net/http"

type noPaginator struct {
}

// make sure it implements the interface
var _ Paginator = (*noPaginator)(nil)

func (noPaginator) HasMore(page int, resp *http.Response) (bool, *http.Request) {
	return false, nil
}
