package generate

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

const (
	AtField      = "at"
	VersionField = "version"
	IDField      = "id"
)

// AllFiles ...
func AllFiles(descriptions []*descriptor.FileDescriptorProto) ([]*plugin_go.CodeGeneratorResponse_File, error) {

	files := []*plugin_go.CodeGeneratorResponse_File{}
	for _, description := range descriptions {
		file, err := File(description)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

// File ...
func File(description *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {

	packageName := description.GetPackage()
	file := jen.NewFile(description.GetName()[len(packageName+"."):])

	for _, message := range description.MessageType {
		switch {
		case isEvent(message):
			file.Add(generateEventMethods(packageName, message))
		case isContainer(message):
		}
	}

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    String(description.GetName()[len(packageName+"."):] + ".es.go"),
		Content: String(file.GoString()),
	}, nil
}

// String returns a reference of a string
func String(s string) *string {
	return &s
}

func fieldMap(fields []*descriptor.FieldDescriptorProto) map[string]*descriptor.FieldDescriptorProto {

	f := make(map[string]*descriptor.FieldDescriptorProto, len(fields))
	for _, field := range fields {
		gogoproto.GetCustomName(field)
		f[strings.ToLower(name(field))] = field
	}

	return f
}

func name(field *descriptor.FieldDescriptorProto) string {
	if gogoproto.IsCustomName(field) {
		return gogoproto.GetCustomName(field)
	}

	return generator.CamelCase(field.GetName())
}

func isContainer(message *descriptor.DescriptorProto) bool {
	for _, oneof := range message.OneofDecl {
		if generator.CamelCase(oneof.GetName()) == "Event" {
			return true
		}
	}

	return false
}

func isEvent(message *descriptor.DescriptorProto) bool {
	var (
		hasAt,
		hasID,
		hasVersion bool
	)

	for _, field := range message.Field {
		switch strings.ToLower(name(field)) {
		case AtField:
			if field.GetType() == descriptor.FieldDescriptorProto_TYPE_INT64 {
				hasAt = true
			}
		case IDField:
			if field.GetType() == descriptor.FieldDescriptorProto_TYPE_STRING {
				hasID = true
			}
		case VersionField:
			if field.GetType() == descriptor.FieldDescriptorProto_TYPE_INT32 {
				hasVersion = true
			}
		}
	}

	return hasAt && hasID && hasVersion
}
