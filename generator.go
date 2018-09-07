package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	pbannotations "google.golang.org/genproto/googleapis/api/annotations"
)

type generator struct{}

func (g *generator) Generate(req *plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	resp := new(plugin.CodeGeneratorResponse)

	for _, name := range req.FileToGenerate {
		file, err := getFileDescriptor(req, name)
		if err != nil {
			return nil, err
		}

		genFile, err := g.generateFile(file)
		if err != nil {
			return nil, fmt.Errorf("generating %s: %s", name, err)
		}

		if len(file.GetService()) > 0 {
			resp.File = append(resp.File, genFile)
		}
	}

	return resp, nil
}

func (g *generator) generateFile(file *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	buffer := new(bytes.Buffer)

	tmpl, err := template.New("file").Parse(browserTemplate)
	if err != nil {
		return nil, err
	}

	data := &templateData{
		Proto:          file,
		Version:        Version,
		SourceFilename: file.GetName(),
		Methods:        map[string]string{},
		URLs:           map[string]string{},
	}

	for _, service := range file.GetService() {
		for _, method := range service.GetMethod() {
			if method.GetOptions() == nil {
				continue
			}

			if !proto.HasExtension(method.Options, pbannotations.E_Http) {
				continue
			}
			ext, err := proto.GetExtension(method.Options, pbannotations.E_Http)
			if err != nil {
				return nil, err
			}

			opts, ok := ext.(*pbannotations.HttpRule)
			if !ok {
				return nil, fmt.Errorf("extension is %T; want an HttpRule", ext)
			}

			var path string
			switch x := opts.GetPattern().(type) {
			case *pbannotations.HttpRule_Get:
				data.Methods[method.GetName()] = "GET"

				rule := opts.GetPattern().(*pbannotations.HttpRule_Get)
				path = rule.Get

			case *pbannotations.HttpRule_Put:
				data.Methods[method.GetName()] = "PUT"

				rule := opts.GetPattern().(*pbannotations.HttpRule_Put)
				path = rule.Put

			case *pbannotations.HttpRule_Post:
				data.Methods[method.GetName()] = "POST"

				rule := opts.GetPattern().(*pbannotations.HttpRule_Post)
				path = rule.Post

			case *pbannotations.HttpRule_Delete:
				data.Methods[method.GetName()] = "DELETE"

				rule := opts.GetPattern().(*pbannotations.HttpRule_Delete)
				path = rule.Delete

			default:
				return nil, fmt.Errorf("unexpected pattern %T", x)
			}

			pathParts := []string{}
			for _, part := range strings.Split(path, "/") {
				if !strings.HasPrefix(part, "{") {
					pathParts = append(pathParts, part)
					continue
				}

				paramParts := []string{"${req."}
				var param string
				for i, c := range part {
					s := string(c)
					if s == "_" || s == "{" || s == "}" {
						continue
					}

					if i > 1 && string(part[i-1]) == "_" {
						param = fmt.Sprintf("%s%s", param, strings.ToUpper(s))
					} else {
						param = fmt.Sprintf("%s%s", param, s)
					}
				}
				paramParts = append(paramParts, param)
				paramParts = append(paramParts, "}")

				pathParts = append(pathParts, strings.Join(paramParts, ""))
			}

			data.URLs[method.GetName()] = strings.Join(pathParts, "/")
		}
	}
	if err = tmpl.Execute(buffer, data); err != nil {
		return nil, fmt.Errorf("rendering template for %s: %s", file.GetName(), err)
	}

	name := file.GetName()
	name = name[:len(name)-len(filepath.Ext(name))]
	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(fmt.Sprintf("%s.js", name)),
		Content: proto.String(buffer.String()),
	}, nil
}

func getFileDescriptor(req *plugin.CodeGeneratorRequest, name string) (*descriptor.FileDescriptorProto, error) {
	for _, descriptor := range req.ProtoFile {
		if descriptor.GetName() == name {
			return descriptor, nil
		}
	}

	return nil, fmt.Errorf("could not find descriptor for %s", name)
}
