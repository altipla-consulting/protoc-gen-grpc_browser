package main

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/reiver/go-stringcase"
	pbannotations "google.golang.org/genproto/googleapis/api/annotations"
)

type templateData struct {
	Version        string
	SourceFilename string
	Services       []*Service
}

func (data *templateData) Quote() string {
	return "`"
}

type Service struct {
	*descriptor.ServiceDescriptorProto
}

func (s *Service) Methods() []*Method {
	methods := []*Method{}
	for _, method := range s.GetMethod() {
		methods = append(methods, &Method{method})
	}

	return methods
}

type Method struct {
	*descriptor.MethodDescriptorProto
}

func (method *Method) HTTPMethod() string {
	if method.GetOptions() == nil {
		return ""
	}

	if !proto.HasExtension(method.Options, pbannotations.E_Http) {
		return ""
	}
	ext, err := proto.GetExtension(method.Options, pbannotations.E_Http)
	if err != nil {
		panic(err)
	}
	opts := ext.(*pbannotations.HttpRule)

	switch opts.GetPattern().(type) {
	case *pbannotations.HttpRule_Get:
		return "GET"

	case *pbannotations.HttpRule_Put:
		return "PUT"

	case *pbannotations.HttpRule_Post:
		return "POST"

	case *pbannotations.HttpRule_Delete:
		return "DELETE"
	}

	return ""
}

func (method *Method) Path() string {
	ext, err := proto.GetExtension(method.Options, pbannotations.E_Http)
	if err != nil {
		panic(err)
	}
	opts := ext.(*pbannotations.HttpRule)

	switch rule := opts.GetPattern().(type) {
	case *pbannotations.HttpRule_Get:
		return rule.Get

	case *pbannotations.HttpRule_Put:
		return rule.Put

	case *pbannotations.HttpRule_Post:
		return rule.Post

	case *pbannotations.HttpRule_Delete:
		return rule.Delete
	}

	panic("http rule has no path")
}

func (method *Method) Binding() string {
	segments := []string{}
	for _, segment := range strings.Split(method.Path(), "/") {
		if !strings.HasPrefix(segment, "{") {
			segments = append(segments, segment)
		} else {
			segments = append(segments, "${req."+stringcase.ToCamelCase(segment[1:]))
		}
	}

	return strings.Join(segments, "/")
}

func (method *Method) HasBody() bool {
	ext, err := proto.GetExtension(method.Options, pbannotations.E_Http)
	if err != nil {
		panic(err)
	}
	opts := ext.(*pbannotations.HttpRule)

	return opts.Body == "*"
}

const browserTemplate = `// Code generated by protoc-gen-grpc_browser {{.Version}}, DO NOT EDIT.
// Source: {{.SourceFilename}}

const grpc = require('@altipla/grpc-browser');

{{range .Services}}
module.exports = class {{.GetName}}Client {
  constructor(opts = {}) {
  	this._caller = new grpc.Caller(opts);
  }{{range .Methods}}{{if .HTTPMethod}}

  {{.GetName}}(req) {
  	return this._caller.send('{{.HTTPMethod}}', {{$.Quote}}{{.Binding}}{{$.Quote}}, req, {{.HasBody}}, '{{.Path}}');
  }{{end}}{{end}}
};
{{end}}
`
