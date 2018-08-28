package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

const Version = "1.0.0"

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	req := new(plugin.CodeGeneratorRequest)
	if err := proto.Unmarshal(data, req); err != nil {
		log.Fatal(err)
	}
	if len(req.FileToGenerate) == 0 {
		log.Fatal("no files to generate")
	}

	gen := new(generator)
	resp, err := gen.Generate(req)
	if err != nil {
		log.Fatal(err)
	}

	data, err = proto.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.Write(data); err != nil {
		log.Fatal(err)
	}
}
