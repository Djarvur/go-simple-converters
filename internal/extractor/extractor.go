// Package extractor enumerating the files and extracting the looks-like-enum lines
package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
	"golang.org/x/tools/go/packages"
)

const (
	enumSuffix = "Name"
	testSuffix = "_test.go"
)

var (
	convertRegex = regexp.MustCompile(`^[Cc]onvert(?:er)?_]`)
)

// ConverterDef describes the enum const record with all the details.
type ConverterDef struct {
	Name    string
	Package string
	Dir     string
	Test    bool

	InType  Type
	OutType Type
}

type Type struct {
	Pkg    string
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type Type
}

// Extract the enum constants relative records.
func Extract(pkgs []*packages.Package, fset *token.FileSet) []ConverterDef {
	res := make([]ConverterDef, 0, 128)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			var (
				fileName = fset.File(file.Pos()).Name()
				dirName  = path.Dir(fileName)
				isTest   = strings.HasSuffix(fileName, testSuffix)
			)

			for _, decl := range file.Decls {
				for _, v := range extractConvertVar(decl) {
					_ = append(
						res,
						ConverterDef{
							Name: v.name,
							Dir:  dirName,
							Test: isTest,
						},
					)
				}
			}
		}
	}

	return res
}

type converter struct {
	name string
	in   Type
	out  Type
}

func extractConvertVar(raw ast.Decl) []converter {
	var res []converter

	decl, ok := raw.(*ast.GenDecl)
	if !ok {
		return nil
	}

	if decl.Tok != token.VAR {
		return nil
	}

	for _, rawSpec := range decl.Specs {
		if v, parsed := parseSpec(rawSpec); parsed {
			slog.Info("extractConvertVar", "parsed", v)
			res = append(res, v)
		}
	}

	return res
}

func parseSpec(raw ast.Spec) (converter, bool) {
	spec, isValue := raw.(*ast.ValueSpec)
	if !isValue || len(spec.Names) < 1 {
		return converter{}, false //nolint:exhaustruct
	}

	funcDecl, ok := spec.Type.(*ast.FuncType)
	if !ok {
		return converter{}, false
	}

	if funcDecl.Params == nil ||
		len(funcDecl.Params.List) != 1 ||
		funcDecl.Results == nil ||
		len(funcDecl.Results.List) != 1 {
		return converter{}, false
	}

	param, ok := resolveType(funcDecl.Params.List[0].Type)
	if !ok {
		return converter{}, false
	}

	result, ok := resolveType(funcDecl.Results.List[0].Type)
	if !ok {
		return converter{}, false
	}

	return converter{
		in:  param,
		out: result,
	}, true
}

func resolveType(expr ast.Expr) (Type, bool) {
	slog.Info("resolveType", "expr", fmt.Sprintf("%T", expr), "expr", expr)
	typeIdent, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return Type{}, false
	}

	return Type{
		Name: typeIdent.Sel.Name,
		Pkg:  extractTypeName(typeIdent.X),
	}, true
}

func extractTypeName(expr ast.Expr) string {
	decl, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}

	return decl.Name
}
