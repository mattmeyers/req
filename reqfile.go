package req

import (
	"fmt"
	"net/http"
)

type Reqfile struct {
	Request    Request     `yaml:"request"`
	Assertions []Assertion `yaml:"assertions"`
}

type Headers map[string]string

type Request struct {
	HTTPVersion string
	Method      string `yaml:"method"`
	Path        string `yaml:"path"`

	Headers Headers `yaml:"headers"`

	Body string `yaml:"body"`
}

func NewRequest() Request {
	return Request{
		Headers: make(Headers),
	}
}

type Assertion struct {
	Name      string `yaml:"name"`
	Condition string `yaml:"condition"`
	fn        AssertionFunc
}

func (a Assertion) Assert(request *http.Request, response *http.Response) error {
	if !a.fn(request, response) {
		return fmt.Errorf("%s failed assertion", a.Name)
	}

	return nil
}
