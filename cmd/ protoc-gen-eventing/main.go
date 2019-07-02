package main

import (
	"io/ioutil"
	"os"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

func main() {

	data, err := ioutil.ReadAll(os.Stdin)
	checkErr(err)

	req := &plugin_go.CodeGeneratorRequest{}

	err = proto.Unmarshal(data, request)
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}