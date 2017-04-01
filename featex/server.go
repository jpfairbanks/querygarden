package featex

import (
	"net/http"
	"net/url"
)

// Request tells us which query to run and what values the arguments take.
type Request struct {
	Key  string     // The name of the query to run
	Args url.Values // values for the bindvars
}

// ParseRequest pulls the url.Values out of the http.Request
func ParseRequest(r *http.Request) (Request, error) {
	var err error
	key := r.URL.Path[len("query/"):]
	args := r.URL.Query()
	return Request{key[1:], args}, err
}
