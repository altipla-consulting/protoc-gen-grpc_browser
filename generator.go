package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
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
