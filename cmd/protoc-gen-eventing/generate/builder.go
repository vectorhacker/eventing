package generate

import (
	"strings"

	jen "github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

const (
	builder = "Builder"
)

func generateBuilder() jen.Code {
	return jen.
		Type().Id("Clock").Interface(
		jen.Id("Now").Params().Qual("time", "Time"),
	).
		Line().Line().
		Type().Id("realClock").Struct().Line().Line().
		Func().Params(jen.Id("realClock")).Id("Now").Params().Qual("time", "Time").Block(
		jen.Return(jen.Qual("time", "Now").Call()),
	).Line().Line().
		Type().Id("BuilderOption").Func().Params(jen.Op("*").Id(builder)).Line().Line().
		Func().Id("WithClock").Params(jen.Id("c").Id("Clock")).Id("BuilderOption").Block(
		jen.Return(
			jen.Func().Params(jen.Id("b").Op("*").Id(builder)).Block(
				jen.Id("b").Dot("clock").Op("=").Id("c"),
			),
		),
	).Line().Line().
		Type().Id(builder).Struct(
		jen.Id("ID").String(),
		jen.Id("Events").Index().Qual("github.com/vectorhacker/eventing", "Event"),
		jen.Id("Version").Int(),
		jen.Id("clock").Id("Clock"),
	).Line().Line().
		Func().Params(jen.Id("b").Op("*").Id(builder)).Id("nextVersion").Params().Int32().Block(
		jen.Id("b").Dot("Version").Op("++"),
		jen.Return(jen.Int32().Parens(jen.Id("b").Dot("Version"))),
	).Line().Line().
		Func().Id("NewBuilder").Params(jen.Id("id").String(), jen.Id("version").Int(), jen.Id("opts").Op("...").Id("BuilderOption")).Op("*").Id(builder).Block(
		jen.Id("b").Op(":=").Op("&").Id(builder).Values(jen.Dict{
			jen.Id("Version"): jen.Id("version"),
			jen.Id("ID"):      jen.Id("id"),
			jen.Id("clock"):   jen.Id("realClock").Values(jen.Dict{}),
		}),

		jen.For(jen.Id("_").Op(",").Id("opt").Op(":=").Range().Id("opts")).Block(
			jen.Id("opt").Call(jen.Id("b")),
		),

		jen.Return(
			jen.Id("b"),
		),
	)
}

func generateEventBuilder(message *descriptor.DescriptorProto, packageName string) jen.Code {
	params := []jen.Code{}
	values := jen.Dict{}

	for _, field := range message.Field {
		switch strings.ToLower(name(field)) {
		case AtField:
			values[jen.Id(name(field))] = jen.Id("b").Dot("clock").Dot("Now").Call().Dot("Unix").Call()
		case VersionField:
			values[jen.Id(name(field))] = jen.Id("b").Dot("nextVersion").Call()
		case IDField:
			values[jen.Id(name(field))] = jen.Id("b").Dot("ID")
		default:
			values[jen.Id(name(field))] = jen.Id(paramCaseName(field))
			param := jen.Id(paramCaseName(field))

			typeName := strings.ReplaceAll(strings.ReplaceAll(field.GetTypeName(), "."+packageName+".", ""), ".", "_")
			switch field.GetType() {
			case descriptor.FieldDescriptorProto_TYPE_BOOL:
				param = param.Bool()
			case descriptor.FieldDescriptorProto_TYPE_BYTES:
				param = param.Index().Byte()
			case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
				param = param.Float64()
			case descriptor.FieldDescriptorProto_TYPE_SFIXED32, descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32:
				param = param.Int32()
			case descriptor.FieldDescriptorProto_TYPE_SFIXED64, descriptor.FieldDescriptorProto_TYPE_INT64,
				descriptor.FieldDescriptorProto_TYPE_SINT64:
				param = param.Int64()
			case descriptor.FieldDescriptorProto_TYPE_FLOAT:
				param = param.Float32()
			case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
				param = param.Op("*").Id(typeName)
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				param = param.String()
			case descriptor.FieldDescriptorProto_TYPE_UINT32:
				param = param.Uint32()
			case descriptor.FieldDescriptorProto_TYPE_UINT64:
				param = param.Uint64()
			case descriptor.FieldDescriptorProto_TYPE_ENUM:
				param = param.Id(typeName)
			}

			params = append(params, param)
		}
	}
	return jen.Func().Params(jen.Id("b").Op("*").Id(builder)).Id(message.GetName()).Params(params...).
		Block(
			jen.Id("e").Op(":=").Op("&").Id(message.GetName()).Values(values),
			jen.Id("b").Dot("Events").Op("=").Append(jen.Id("b").Dot("Events"), jen.Id("e")),
		)
}
