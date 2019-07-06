package generate

import (
	jen "github.com/dave/jennifer/jen"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func generateEventMethods(message *descriptor.DescriptorProto) jen.Code {

	fields := fieldMap(message.Field)

	atName := name(fields["at"])
	versionName := name(fields["version"])
	idName := name(fields["id"])

	return jen.
		Line().
		Comment("AggregateID implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("AggregateID").
		Params().String().
		Block(
			jen.Return(jen.Id("e").Dot(idName)),
		).Line().
		Comment("EventVersion implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("EventVersion").
		Params().Int().
		Block(
			jen.Return(jen.Int().Params(jen.Id("e").Dot(versionName))),
		).Line().
		Comment("EventAt implements the Event interface").Line().
		Func().Params(jen.Id("e").Op("*").Id(message.GetName())).Id("EventAt").
		Params().Qual("time", "Time").
		Block(
			jen.Return(jen.Qual("time", "Unix").Call(jen.Id("e").Dot(atName), jen.Lit(0))),
		).Line().Line().
		Comment("EventName implements the EventNamer interface").Line().
		Func().Params(jen.Id(message.GetName())).Id("EventName").Params().String().Block(
		jen.Return(jen.Lit(message.GetName())),
	).Line()
}
