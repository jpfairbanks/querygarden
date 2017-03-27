package featex

import (
	"net/http"
	"net/url"
)

type Request struct {
	Key string
	Args url.Values
}

func ParseRequest(r *http.Request) (Request, error) {
	var err error
	key := r.URL.Path[len("query/"):]
	args := r.URL.Query()
	return Request{key[1:], args}, err
}