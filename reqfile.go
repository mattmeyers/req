package reql

import (
	"fmt"
	"net/http"
)

type Reqfile struct {
	Request  Request  `hcl:"request,block"`
	Response Response `hcl:"response,block"`
}

type Headers struct {
	Values map[string]string `hcl:",remain"`
}

type Request struct {
	HTTPVersion string
	Method      string `hcl:"method"`
	URL         string `hcl:"url"`

	Headers map[string]string `hcl:"headers,optional"`

	Body string `hcl:"body,optional"`
}

func NewRequest() Request {
	return Request{
		// Headers: make(Headers),
	}
}

type Response struct {
	Assertions []Assertion `hcl:"assert,block"`
}

type Assertion struct {
	Name string `hcl:"name,label"`
	Expr string `hcl:"expr"`
	fn   AssertionFunc
}

func (a Assertion) Assert(request *http.Request, response *http.Response) error {
	if !a.fn(request, response) {
		return fmt.Errorf("%s failed assertion", a.Name)
	}

	return nil
}
