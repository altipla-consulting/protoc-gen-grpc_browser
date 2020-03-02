package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

const Version = "1"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	req := new(plugin.CodeGeneratorRequest)
	if err := proto.Unmarshal(data, req); err != nil {
		return err
	}
	if len(req.FileToGenerate) == 0 {
		log.Fatal("no files to generate")
	}

	resp := new(plugin.CodeGeneratorResponse)

	for _, name := range req.FileToGenerate {
		file, err := getFileDescriptor(req, name)
		if err != nil {
			return err
		}

		genFile, err := generateFile(file)
		if err != nil {
			return fmt.Errorf("generating %s: %s", name, err)
		}

		if len(file.GetService()) > 0 {
			resp.File = append(resp.File, genFile)
		}
	}

	data, err = proto.Marshal(resp)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}

	return nil
}

func generateFile(file *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	tmpl, err := template.New("file").Parse(browserTemplate)
	if err != nil {
		return nil, err
	}

	data := &templateData{
		Version:        Version,
		SourceFilename: file.GetName(),
	}
	for _, service := range file.GetService() {
		data.Services = append(data.Services, &Service{service})
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering template for %s: %s", file.GetName(), err)
	}

	name := file.GetName()
	name = name[:len(name)-len(filepath.Ext(name))]
	return &plugin.CodeGeneratorResponse_File{
		Name:    proto.String(fmt.Sprintf("%s.js", name)),
		Content: proto.String(buf.String()),
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
