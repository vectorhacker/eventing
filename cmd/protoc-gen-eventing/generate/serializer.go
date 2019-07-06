package generate

import (
	"strings"

	jen "github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

const (
	serializer           = "Serializer"
	eventsourcingPackage = "github.com/vectorhacker/eventing"
)

func holderName(containerMessage *descriptor.DescriptorProto, field *descriptor.FieldDescriptorProto) string {
	return containerMessage.GetName() + "_" + name(field)
}

func generateSerializer(containerMessage *descriptor.DescriptorProto, packageName string) jen.Code {

	strct := jen.Type().Id(serializer).Struct().Line()

	serializeFunc := jen.Line().
		Func().Params(jen.Id("s").Id(serializer)).Id("Serialize").
		Params(
			jen.Id("event").Qual(eventsourcingPackage, "Event"),
		).Params(jen.Qual(eventsourcingPackage, "Record"), jen.Error())

	deserializeFunc := jen.Line().
		Func().Params(jen.Id("s").Id(serializer)).Id("Deserialize").
		Params(
			jen.Id("record").Qual(eventsourcingPackage, "Record"),
		).Params(jen.Qual(eventsourcingPackage, "Event"), jen.Error())

	eventSerializeSwitch := jen.Switch(jen.Id("e").Op(":=").Id("event").Op(".").Parens(jen.Type()))
	eventDeserializeSwitch := jen.Switch(jen.Id("container").Dot("Event").Op(".").Parens(jen.Type()))

	serializeCases := []jen.Code{}
	deserializeCases := []jen.Code{}
	for _, field := range containerMessage.Field {
		if field.OneofIndex == nil {
			continue
		}

		oenof := containerMessage.OneofDecl[field.GetOneofIndex()]
		if oenof.GetName() != "event" {
			continue
		}

		serializeCases = append(serializeCases, jen.Case(jen.Op("*").Id(typeName(field, packageName))).Block(
			jen.Id("container").Dot("Event").Op("=").Op("&").Id(holderName(containerMessage, field)).Values(jen.Dict{
				jen.Id(name(field)): jen.Id("e"),
			}),
		))

		deserializeCases = append(deserializeCases, jen.Case(jen.Op("*").Id(holderName(containerMessage, field))).
			Block(
				jen.Return(jen.Id("container").Dot("Get"+name(field)).Call(), jen.Nil()),
			))
	}

	_ = deserializeFunc

	eventSerializeSwitch = eventSerializeSwitch.Block(serializeCases...)
	eventDeserializeSwitch = eventDeserializeSwitch.Block(deserializeCases...)

	serializeFunc = serializeFunc.Block(
		jen.Id("container").Op(":=").Op("&").Id(containerMessage.GetName()).Values(jen.Dict{}),
		eventSerializeSwitch.Line(),
		jen.List(jen.Id("data"), jen.Err()).Op(":=").Qual("github.com/gogo/protobuf/proto", "Marshal").Call(jen.Id("container")).Line(),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.List(jen.Qual(eventsourcingPackage, "Record").Values(jen.Dict{}), jen.Err())),
		).Line(),
		jen.Return(jen.Qual(eventsourcingPackage, "Record").Values(jen.Dict{
			jen.Id("Data"):    jen.Id("data"),
			jen.Id("Version"): jen.Id("event").Dot("EventVersion").Call(),
		}), jen.Nil()),
	)

	deserializeFunc = deserializeFunc.Block(
		jen.Id("container").Op(":=").Op("&").Id(containerMessage.GetName()).Values(),
		jen.Err().Op(":=").Qual("github.com/gogo/protobuf/proto", "Unmarshal").Call(jen.Id("record").Dot("Data"), jen.Id("container")),
		jen.If(jen.Err().Op("!=").Nil()).Block(jen.Return(jen.Nil(), jen.Err())),
		eventDeserializeSwitch.Line(),
		jen.Return(jen.Nil(), jen.Qual("errors", "New").Call(jen.Lit("No event"))),
	)

	return jen.Add(strct, jen.Func().Id("NewSerializer").Params().Qual(eventsourcingPackage, "Serializer").Block(
		jen.Return(jen.Op("&").Id(serializer).Values()),
	).Line(), serializeFunc.Line(), deserializeFunc.Line())
}

func typeName(field *descriptor.FieldDescriptorProto, packageName string) string {

	return strings.ReplaceAll(field.GetTypeName()[len("."+packageName+"."):], ".", "_")
}
