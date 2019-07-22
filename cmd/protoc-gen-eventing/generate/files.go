package generate

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	plugin_go "github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

// Field constants
const (
	AtField      = "at"
	VersionField = "version"
	IDField      = "id"
)

// AllFiles generates the code for all files.
func AllFiles(descriptions []*descriptor.FileDescriptorProto) ([]*plugin_go.CodeGeneratorResponse_File, error) {

	files := []*plugin_go.CodeGeneratorResponse_File{}
	for _, description := range descriptions {
		if !hasEvents(description) {
			continue
		}
		file, err := File(description)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

// File generates the events and stuff for this protobuf description
func File(description *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {

	packageName := description.GetPackage()
	file := jen.NewFile(packageName)

	file.Add(generateBuilder())

	for _, message := range description.MessageType {
		switch {
		case isEvent(message):
			file.Add(generateEventMethods(message))
			file.Add(generateEventBuilder(message, packageName))
		case isContainer(message):
			file.Add(generateSerializer(message, packageName))
		}
	}

	name := strings.ReplaceAll(description.GetName(), ".proto", ".es.go")

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    String(name),
		Content: String(file.GoString()),
	}, nil
}

// String returns a reference of a string
func String(s string) *string {
	return &s
}

func hasEvents(description *descriptor.FileDescriptorProto) bool {
	for _, message := range description.MessageType {
		if isEvent(message) {
			return true
		}
	}

	return false
}

func fieldMap(fields []*descriptor.FieldDescriptorProto) map[string]*descriptor.FieldDescriptorProto {

	f := make(map[string]*descriptor.FieldDescriptorProto, len(fields))
	for _, field := range fields {
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

func isRestrictedName(name string) bool {

	switch name {
	case "type",
		"struct",
		"interface",
		"map",
		"append",
		"make",
		"new",
		"func",
		"return",
		"for",
		"switch",
		"case",
		"if",
		"else",
		"var",
		"const",
		"package",
		"import":
		return true
	}

	return false
}

func paramCaseName(field *descriptor.FieldDescriptorProto) string {
	gotName := func() string {
		n := name(field)

		letters := strings.Split(n, "")
		letters[0] = strings.ToLower(letters[0])

		return strings.Join(letters, "")
	}()

	if isRestrictedName(gotName) {
		return "_" + gotName
	}

	return gotName
}

func isContainer(message *descriptor.DescriptorProto) bool {
	for _, oneof := range message.OneofDecl {
		if generator.CamelCase(oneof.GetName()) == "Event" {
			return true
		}
	}

	return false
}

// isEvent determines if a message is an event if it has the at, version, and id fields
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
