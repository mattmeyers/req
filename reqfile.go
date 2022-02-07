package req

type Reqfile struct {
	Request Request `yaml:"request"`
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
	Name       string
	Statements []string
}
