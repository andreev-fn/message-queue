package e2eutils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pb33f/libopenapi"
	oapivalidator "github.com/pb33f/libopenapi-validator"

	"server/internal/openapi"
)

var validator oapivalidator.Validator

func init() {
	doc, err := libopenapi.NewDocument(openapi.GetSpecYAML())
	if err != nil {
		panic(err)
	}

	v, errs := oapivalidator.NewValidator(doc)
	if len(errs) > 0 {
		panic(errs[0])
	}

	validator = v
}

type HTTPTestDoer struct {
	t       *testing.T
	handler http.Handler
}

func NewHTTPTestDoer(t *testing.T, handler http.Handler) HTTPTestDoer {
	return HTTPTestDoer{t: t, handler: handler}
}

func (d HTTPTestDoer) Do(req *http.Request) (*http.Response, error) {
	reqValid, validationErrors := validator.ValidateHttpRequest(req)
	if !reqValid {
		for i := range validationErrors {
			d.t.Error(validationErrors[i].Error())
		}
	}

	recorder := httptest.NewRecorder()
	d.handler.ServeHTTP(recorder, req)
	resp := recorder.Result()

	respValid, validationErrors := validator.ValidateHttpResponse(req, resp)
	if !respValid {
		for i := range validationErrors {
			d.t.Error(validationErrors[i].Error())
		}
	}

	if resp.StatusCode == http.StatusInternalServerError {
		d.t.Error("internal server error not expected in tests")
	}

	return resp, nil
}
