package req

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
)

func ParseReqfile(path string, env map[string]string) (Reqfile, error) {
	vars := map[string]cty.Value{"env": cty.MapVal(map[string]cty.Value{"": cty.StringVal("")})}

	if len(env) > 0 {
		envMap := map[string]cty.Value{}
		for k, v := range env {
			envMap[k] = cty.StringVal(v)
		}

		vars["env"] = cty.MapVal(envMap)
	}

	var reqfile Reqfile
	err := hclsimple.DecodeFile(
		path,
		&hcl.EvalContext{Variables: vars},
		&reqfile,
	)
	if err != nil {
		return Reqfile{}, err
	}

	for i := range reqfile.Response.Assertions {
		reqfile.Response.Assertions[i].fn = ParseAssertion(reqfile.Response.Assertions[i].Expr)
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
