package generator

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	plugin_go "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/printer"
	"log"
	"os"
	"path"
	"strings"
)

// Generator is responsible
type Generator struct {
	*generator.Generator // github.com/golang/protobuf/protoc-gen-go/generator

	qualifiedNameStack StringStack
	qualifiedTypeMap   map[string]string // Map qualified proto3 types and names to GraphQL names
}

// New creates a new default initialized Generator.
func New() *Generator {
	return &Generator{
		Generator: generator.New(),

		qualifiedNameStack: make(StringStack, 0),
		qualifiedTypeMap: map[string]string{
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_INT32)]:    "Int",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_SFIXED32)]: "Int",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_SINT32)]:   "Int",

			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_INT64)]:    "Int64",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_SFIXED64)]: "Int64",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_SINT64)]:   "Int64",

			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_FIXED32)]: "Uint32",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_FIXED64)]: "Uint64",

			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_FLOAT)]:  "Float",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_DOUBLE)]: "Float64",

			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_BYTES)]:  "String",
			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]: "String",

			descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_BOOL)]: "Boolean",
		},
	}
}

func (g *Generator) Error(err error, msgs ...string) {
	s := strings.Join(msgs, " ") + ":" + err.Error()
	log.Print("protoc-gen-graphql: error:", s)
	os.Exit(1)
}

// Fail with proper annotation
func (g *Generator) Fail(msgs ...string) {
	s := strings.Join(msgs, " ")
	log.Print("protoc-gen-graphql: error:", s)
	os.Exit(1)
}

// GenerateAllFiles generates all GraphQL schema files from the input protobuf files.
func (g *Generator) GenerateAllFiles() {
	g.Generator.GenerateAllFiles()
	g.Generator.Reset()
	g.Generator.Response = new(plugin_go.CodeGeneratorResponse)

	genFileMap := make(map[string]bool, len(g.Generator.Request.FileToGenerate))
	for _, fileName := range g.Generator.Request.FileToGenerate {
		genFileMap[fileName] = true
	}

	for _, protoFile := range g.Generator.Request.ProtoFile {
		if _, ok := genFileMap[protoFile.GetName()]; !ok {
			continue
		}
		doc := ast.NewDocument(nil)

		g.qualifiedNameStack = g.qualifiedNameStack.Push(*protoFile.Package)

		// EnumType
		for _, enum := range protoFile.EnumType {
			doc.Definitions = append(doc.Definitions, g.addGraphQLEnum(enum))
		}

		// MessageType
		for _, message := range protoFile.MessageType {
			doc.Definitions = append(doc.Definitions, g.addGraphQLType(message)...)
		}

		// Service

		data, _ := printer.Print(doc).(string)
		g.Response.File = append(g.Response.File, &plugin_go.CodeGeneratorResponse_File{
			Name:    proto.String(g.graphqlFileName(*protoFile.Name)),
			Content: proto.String(data),
		})
	}
}

func (g *Generator) addGraphQLEnum(enum *descriptor.EnumDescriptorProto) ast.Node {
	s := g.qualifiedNameStack.Push(*enum.Name)
	graphqlName := strings.Join(s[1:], "")
	g.qualifiedTypeMap["."+strings.Join(s, ".")] = graphqlName

	enumDef := ast.NewEnumDefinition(&ast.EnumDefinition{
		Name: ast.NewName(&ast.Name{Value: graphqlName}),
	})

	for _, enumValue := range enum.Value {
		enumDef.Values = append(enumDef.Values, ast.NewEnumValueDefinition(&ast.EnumValueDefinition{
			Name: ast.NewName(&ast.Name{Value: *enumValue.Name}),
		}))
	}

	return enumDef
}

func (g *Generator) addGraphQLType(message *descriptor.DescriptorProto) []ast.Node {
	g.qualifiedNameStack = g.qualifiedNameStack.Push(*message.Name)
	graphqlName := strings.Join(g.qualifiedNameStack[1:], "")
	g.qualifiedTypeMap["."+strings.Join(g.qualifiedNameStack, ".")] = graphqlName

	objDef := ast.NewObjectDefinition(&ast.ObjectDefinition{
		Name: ast.NewName(&ast.Name{Value: graphqlName}),
	})

	var nodes []ast.Node

	// Enum Types
	for _, enum := range message.EnumType {
		nodes = append(nodes, g.addGraphQLEnum(enum))
	}

	// Nested Types
	for _, nested := range message.NestedType {
		nodes = append(nodes, g.addGraphQLType(nested)...)
	}

	// Fields
	for _, field := range message.Field {
		fieldDef := &ast.FieldDefinition{
			Name: ast.NewName(&ast.Name{Value: *field.Name}),
		}

		// Type
		typeName := ""
		if field.Type != nil && field.TypeName != nil {
			if tmpTypeName, ok := g.qualifiedTypeMap[*field.TypeName]; ok {
				typeName = tmpTypeName
			} else {
				nodes = append(nodes, g.addCustomScalar(*field.TypeName))
				typeName = g.qualifiedTypeMap[*field.TypeName]
			}
		} else if field.TypeName != nil {
			typeName, _ = g.qualifiedTypeMap[*field.TypeName]
			if tmpTypeName, ok := g.qualifiedTypeMap[*field.TypeName]; ok {
				typeName = tmpTypeName
			}
		} else if field.Type != nil {
			if tmpTypeName, ok := g.qualifiedTypeMap[descriptor.FieldDescriptorProto_Type_name[int32(*field.Type)]]; ok {
				typeName = tmpTypeName
			}
		}

		if typeName != "" {
			switch *field.Label {
			case descriptor.FieldDescriptorProto_LABEL_OPTIONAL:
				fieldDef.Type = ast.NewNamed(&ast.Named{
					Name: ast.NewName(&ast.Name{Value: typeName}),
				})

			case descriptor.FieldDescriptorProto_LABEL_REQUIRED:
				fieldDef.Type = ast.NewNonNull(&ast.NonNull{
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: typeName,
						}),
					}),
				})

			case descriptor.FieldDescriptorProto_LABEL_REPEATED:
				fieldDef.Type = ast.NewList(&ast.List{
					Type: ast.NewNamed(&ast.Named{
						Name: ast.NewName(&ast.Name{
							Value: typeName,
						}),
					}),
				})
			}
		}

		objDef.Fields = append(objDef.Fields, ast.NewFieldDefinition(fieldDef))
	}

	g.qualifiedNameStack, _, _ = g.qualifiedNameStack.Pop()

	return append(nodes, objDef)

}

func (g *Generator) addCustomScalar(name string) ast.Node {
	graphqlName := g.qualifiedNameToGraphQLName(name)
	g.qualifiedTypeMap[name] = graphqlName

	return ast.NewScalarDefinition(&ast.ScalarDefinition{
		Name: ast.NewName(&ast.Name{
			Value: graphqlName,
		}),
	})
}

func (g *Generator) qualifiedNameToGraphQLName(name string) string {
	return strings.ReplaceAll(strings.Title(name), ".", "")

}

func (g *Generator) graphqlFileName(name string) string {
	if ext := path.Ext(name); ext == ".proto" || ext == ".protodevel" {
		name = name[:len(name)-len(ext)]
	}
	name += ".pb.graphql"

	return name
}
