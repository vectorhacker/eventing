package generate

import (
	jen "github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func generateEventMethods(
	packageName string,
	message *descriptor.DescriptorProto,
) jen.Code {

	fields := fieldMap(message.Field)

	atName := name(fields["at"])
	versionName := name(fields["version"])
	idName := name(fields["id"])

	return jen.
		Comment("AggregateID implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("AggregateID").
		Params().
		Block(
			jen.Return(jen.Id(idName)),
		).Line().
		Comment("EventVersion implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("EventVersion").
		Params().
		Block(
			jen.Return(jen.Int().Params(jen.Id(versionName))),
		).Line().
		Comment("EventAt implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("EventAt").
		Params().
		Block(
			jen.Return(jen.Qual("time", "Unix").Call(jen.Id(atName), jen.Lit(0))),
		).Line()
}
