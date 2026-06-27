package schemaindex_test

import (
	"fmt"

	"github.com/brian-bell/graphql-parser/ast"
	"github.com/brian-bell/graphql-parser/parser"
	"github.com/brian-bell/graphql-parser/schemaindex"
)

func ExampleIndex_LookupType() {
	doc, err := parser.ParseSchema(`
		type Book { title: String }
		extend type Book { isbn: ID }
	`)
	if err != nil {
		panic(err)
	}

	idx := schemaindex.New(doc)
	book := idx.LookupType("Book")
	base := book.BaseDefinitions()[0].(*ast.ObjectTypeDefinition)
	ext := book.Extensions()[0].(*ast.ObjectTypeExtension)

	fmt.Println(book.Name())
	fmt.Println(len(book.BaseDefinitions()))
	fmt.Println(len(book.Extensions()))
	fmt.Printf("%T %s\n", base, base.Name)
	fmt.Printf("%T %s\n", ext, ext.Fields[0].Name)

	// Output:
	// Book
	// 1
	// 1
	// *ast.ObjectTypeDefinition Book
	// *ast.ObjectTypeExtension isbn
}
