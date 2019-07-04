package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/gogo/protobuf/proto"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/vectorhacker/eventing/cmd/protoc-gen-eventing/generate"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	checkErr(err)

	request := &plugin_go.CodeGeneratorRequest{}
	checkErr(proto.Unmarshal(data, request))

	files, err := generate.AllFiles(request.ProtoFile)
	checkErr(err)

	response := &plugin_go.CodeGeneratorResponse{
		File: files,
	}

	data, err = proto.Marshal(response)
	checkErr(err)

	_, err = os.Stdout.Write(data)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}
