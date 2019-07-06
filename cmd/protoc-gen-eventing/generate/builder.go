package generate

import (
	jen "github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

const (
	builder = "Builder"
)

func generateBuilder() jen.Code {
	return jen.Type().Id(builder).Struct(
		jen.Id("ID").String(),
		jen.Id("Events").Index().Qual("github.com/vectorhacker/eventing", "Event"),
		jen.Id("Version").Int(),
	).Line().Line().
		Func().Params(jen.Id("b").Op("*").Id(builder)).Id("nextVersion").Params().Block(
		jen.Id("b").Dot("Version").Op("++"),
	).Line().Line().
		Func().Id("NewBuilder").Params(jen.Id("id").String(), jen.Id("version").Int()).Op("*").Id(builder).Block(
		jen.Return(
			jen.Op("&").Id(builder).Values(jen.Dict{
				jen.Id("Version"): jen.Id("version"),
				jen.Id("ID"):      jen.Id("id"),
			}),
		),
	)
}

func generateEventBuilder(message *descriptor.DescriptorProto, packageName string) jen.Code {

	fields := fieldMap(message.Field)

	params := []jen.Code{}
	values := jen.Dict{}

	for fieldName, field := range fields {
		switch fieldName {
		case AtField:
			values[jen.Id(name(field))] = jen.Qual("time", "Now()").Dot("Unix").Call()
		case VersionField:
			values[jen.Id(name(field))] = jen.Int32().Parens(jen.Id("b").Dot("Version"))
		case IDField:
			values[jen.Id(name(field))] = jen.Id("b").Dot("ID")
		default:
			values[jen.Id(name(field))] = jen.Id(paramCaseName(field))
			param := jen.Id(paramCaseName(field))

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
				param = param.Op("*").Id(field.GetTypeName()[len("."+packageName+"."):])
			case descriptor.FieldDescriptorProto_TYPE_STRING:
				param = param.String()
			case descriptor.FieldDescriptorProto_TYPE_UINT32:
				param = param.Uint32()
			case descriptor.FieldDescriptorProto_TYPE_UINT64:
				param = param.Uint64()
			case descriptor.FieldDescriptorProto_TYPE_ENUM:
				param = param.Id(field.GetTypeName()[len("."+packageName+"."):])
			}

			params = append(params, param)
		}
	}
	return jen.Func().Params(jen.Id("b").Op("*").Id(builder)).Id(message.GetName()).Params(params...).
		Block(
			jen.Id("e").Op(":=").Op("&").Id(message.GetName()).Values(values),
			jen.Id("b").Dot("Events").Op("=").Append(jen.Id("b").Dot("Events"), jen.Id("e")),
			jen.Id("b").Dot("nextVersion").Call(),
		)
}
