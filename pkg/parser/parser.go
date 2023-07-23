package parser

import (
	"bytes"
	"io"
	"strings"

	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/jsonpath"
)

func Decode(reader io.Reader) (runtime.Object, string, error) {
	rawData, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}

	decoder := scheme.Codecs.UniversalDecoder(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	data, gvk, err := decoder.Decode(rawData, nil, nil)
	if err != nil {
		return nil, "", err
	}
	return data, gvk.Kind, nil
}

func Allowed(target string, allowedKinds []string) bool {
	for i := 0; i < len(allowedKinds); i++ {
		if allowedKinds[i] == "*" {
			return true
		}
		allowedKinds[i] = strings.ToLower(allowedKinds[i])
	}
	return slices.Contains(allowedKinds, strings.ToLower(target))
}

func RenderTemplates(object runtime.Object, templates []string) ([]string, error) {
	var resp []string
	for _, tpl := range templates {
		p, err := printers.NewJSONPathPrinter(strings.TrimSpace(tpl))
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		p.PrintObj(object, buf)
		resp = append(resp, buf.String())

	}
	return resp, nil
}

func Parse(template string, input interface{}) (string, error) {
	j := jsonpath.New("").AllowMissingKeys(true)

	if err := j.Parse(template); err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err := j.Execute(buf, input)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
