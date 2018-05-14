package main

import (
	"bytes"
	"github.com/emicklei/proto"
	"io"
)

var BUILTINS = map[string]string{
	"double":   "Float",
	"float":    "Float",
	"int32":    "Int",
	"int64":    "Int",
	"uint32":   "Int",
	"uint64":   "Int",
	"sint32":   "Int",
	"sint64":   "Int",
	"fixed32":  "Int",
	"fixed64":  "Int",
	"sfixed32": "Int",
	"sfixed64": "Int",
	"bool":     "Boolean",
	"string":   "String",
	"bytes":    "[String]",
}

type (
	Visitor struct {
		scope    *Scope
		buff     *bytes.Buffer
		children []*Visitor
	}
)

func NewVisitor(converter *Converter) *Visitor {
	return &Visitor{
		buff:     new(bytes.Buffer),
		children: make([]*Visitor, 0, 5),
		scope:    NewScope(converter),
	}
}

func (v *Visitor) Fork(name string) *Visitor {
	child := &Visitor{
		buff:     new(bytes.Buffer),
		children: make([]*Visitor, 0, 5),
		scope:    v.scope.Fork(name),
	}

	v.children = append(v.children, child)

	return child
}

func (v *Visitor) Flush(out io.Writer) {
	out.Write(v.buff.Bytes())

	v.buff.Reset()

	for _, child := range v.children {
		child.Flush(out)
	}
}

func (v *Visitor) VisitMessage(m *proto.Message) {
	v.buff.WriteString("\n")

	v.scope.AddLocalType(m.Name)

	v.buff.WriteString("type " + v.scope.converter.NewTypeName(v.scope, m.Name) + " {\n")

	fields := make([]*proto.NormalField, 0, len(m.Elements))

	for _, element := range m.Elements {

		field, ok := element.(*proto.NormalField)

		// it's not a nested message/enum
		if ok == true {
			// we put it in array in order to process nested messages first
			// in case they exist and have them in a scope
			fields = append(fields, field)
		} else {
			// if so, create a nested visitor
			// we need to track a parent's name
			// in order to generate a unique name for nested ones
			// we create another visitor in order to unfold nested types since GraphQL does not support nested types
			element.Accept(v.Fork(m.Name))
		}
	}

	// now, having all nested messages in a scope, we can transform fields
	for _, field := range fields {
		field.Accept(v)
	}

	v.buff.WriteString("}\n")

}
func (v *Visitor) VisitService(s *proto.Service) {}
func (v *Visitor) VisitSyntax(s *proto.Syntax)   {}
func (v *Visitor) VisitPackage(p *proto.Package) {
	v.scope.SetPackageName(p.Name)
}
func (v *Visitor) VisitOption(o *proto.Option) {}
func (v *Visitor) VisitImport(i *proto.Import) {
	v.scope.AddImportedType(i.Filename)
}
func (v *Visitor) VisitNormalField(field *proto.NormalField) {
	v.buff.WriteString("    " + field.Name + ":")

	typeName := v.scope.ResolveTypeName(field.Type)

	if field.Repeated == false {
		v.buff.WriteString(" " + typeName)
	} else {
		v.buff.WriteString(" [" + typeName + "]")
	}

	if field.Required == true {
		v.buff.WriteString("!")
	}

	v.buff.WriteString("\n")
}
func (v *Visitor) VisitEnumField(i *proto.EnumField) {
	v.buff.WriteString("    " + i.Name + "\n")
}
func (v *Visitor) VisitEnum(e *proto.Enum) {
	v.scope.AddLocalType(e.Name)

	v.buff.WriteString("\n")

	v.buff.WriteString("enum " + v.scope.converter.NewTypeName(v.scope, e.Name) + " {\n")

	for _, element := range e.Elements {
		element.Accept(v)
	}

	v.buff.WriteString("}\n")
}
func (v *Visitor) VisitComment(e *proto.Comment)       {}
func (v *Visitor) VisitOneof(o *proto.Oneof)           {}
func (v *Visitor) VisitOneofField(o *proto.OneOfField) {}
func (v *Visitor) VisitReserved(r *proto.Reserved)     {}
func (v *Visitor) VisitRPC(r *proto.RPC)               {}
func (v *Visitor) VisitMapField(f *proto.MapField)     {}

// proto2
func (v *Visitor) VisitGroup(g *proto.Group)           {}
func (v *Visitor) VisitExtensions(e *proto.Extensions) {}
