package req

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func ParseReqfile(path string) (Reqfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return Reqfile{}, err
	}

	var reqfile Reqfile
	err = yaml.NewDecoder(f).Decode(&reqfile)
	if err != nil {
		return Reqfile{}, err
	}

	for i := range reqfile.Assertions {
		reqfile.Assertions[i].fn = ParseAssertion(reqfile.Assertions[i].Condition)
	}

	return reqfile, nil
}

type AssertionFunc func(*http.Request, *http.Response) bool

func ParseAssertion(cond string) AssertionFunc {
	return func(request *http.Request, response *http.Response) bool {
		parts := strings.SplitN(cond, " ", 3)
		var l, r string
		if strings.HasPrefix(parts[0], "res") {
			l = responseProperty(response, parts[0][(strings.Index(parts[0], ".")+1):])
		} else {
			panic("idk what to do")
		}

		r = parts[2]

		fmt.Printf("Asserting %s %s %s\n", l, parts[1], r)

		return getComparator(parts[1])(l, r)
	}
}

func responseProperty(res *http.Response, property string) string {
	switch {
	case property == "code":
		return strconv.Itoa(res.StatusCode)
	case strings.HasPrefix(property, "headers"):
		return res.Header.Get(property[(strings.Index(property, ".") + 1):])
	case property == "body":
		b, _ := io.ReadAll(res.Body)
		return string(b)
	}

	return ""
}

func getComparator(s string) func(string, string) bool {
	switch s {
	case "==":
		return func(s1, s2 string) bool { return s1 == s2 }
	case "!=":
		return func(s1, s2 string) bool { return s1 != s2 }
	case ">":
		return func(s1, s2 string) bool { return s1 > s2 }
	case ">=":
		return func(s1, s2 string) bool { return s1 >= s2 }
	case "<":
		return func(s1, s2 string) bool { return s1 < s2 }
	case "<=":
		return func(s1, s2 string) bool { return s1 <= s2 }
	default:
		panic("unknown comparator")
	}
}
