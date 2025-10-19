package e2eutils

import (
	"net/http"
	"net/http/httptest"
)

type HTTPTestDoer struct {
	handler http.Handler
}

func NewHTTPTestDoer(handler http.Handler) HTTPTestDoer {
	return HTTPTestDoer{handler: handler}
}

func (d HTTPTestDoer) Do(req *http.Request) (*http.Response, error) {
	resp := httptest.NewRecorder()
	d.handler.ServeHTTP(resp, req)
	return resp.Result(), nil
}
