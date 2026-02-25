package parser

import (
	"bytes"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
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
	for _, kind := range allowedKinds {
		if kind == "*" {
			return true
		}
		if strings.EqualFold(kind, target) {
			return true
		}
	}
	return false
}

func RenderTemplates(object runtime.Object, templates []string) ([]string, error) {
	var resp []string
	for _, tpl := range templates {
		p, err := printers.NewJSONPathPrinter(strings.TrimSpace(tpl))
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		if err := p.PrintObj(object, buf); err != nil {
			return nil, err
		}
		resp = append(resp, buf.String())

	}
	return resp, nil
}
