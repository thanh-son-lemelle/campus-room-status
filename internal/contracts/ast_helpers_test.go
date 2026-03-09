package contracts

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type parsedPackage struct {
	fset      *token.FileSet
	files     []*ast.File
	typeSpecs map[string]*ast.TypeSpec
}

func parsePackage(t *testing.T, relDir string) parsedPackage {
	t.Helper()

	absDir, err := filepath.Abs(relDir)
	if err != nil {
		t.Fatalf("failed to resolve %q: %v", relDir, err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		t.Fatalf("expected package directory %q to exist: %v", relDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", relDir)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(
		fset,
		absDir,
		func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") },
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf("failed to parse package in %q: %v", relDir, err)
	}
	if len(pkgs) == 0 {
		t.Fatalf("no package found in %q", relDir)
	}

	var selected *ast.Package
	for _, pkg := range pkgs {
		selected = pkg
		break
	}

	files := make([]*ast.File, 0, len(selected.Files))
	typeSpecs := make(map[string]*ast.TypeSpec)

	for _, file := range selected.Files {
		files = append(files, file)

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				typeSpecs[typeSpec.Name.Name] = typeSpec
			}
		}
	}

	return parsedPackage{
		fset:      fset,
		files:     files,
		typeSpecs: typeSpecs,
	}
}

func requireStructFields(t *testing.T, pkg parsedPackage, typeName string, expected map[string]string) map[string]*ast.Field {
	t.Helper()

	typeSpec, ok := pkg.typeSpecs[typeName]
	if !ok {
		t.Fatalf("expected type %q to exist", typeName)
	}

	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		t.Fatalf("expected %q to be a struct", typeName)
	}

	fieldsByName := make(map[string]*ast.Field)
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		for _, name := range field.Names {
			fieldsByName[name.Name] = field
		}
	}

	for fieldName, expectedType := range expected {
		field, ok := fieldsByName[fieldName]
		if !ok {
			t.Fatalf("expected %q.%s field to exist", typeName, fieldName)
		}

		if expectedType == "" {
			continue
		}

		gotType := exprString(pkg.fset, field.Type)
		if gotType != expectedType {
			t.Fatalf("expected %q.%s to be %q, got %q", typeName, fieldName, expectedType, gotType)
		}
	}

	return fieldsByName
}

func requireInterfaceMethodContains(t *testing.T, pkg parsedPackage, typeName string, methodName string, fragments []string) {
	t.Helper()

	typeSpec, ok := pkg.typeSpecs[typeName]
	if !ok {
		t.Fatalf("expected interface %q to exist", typeName)
	}

	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		t.Fatalf("expected %q to be an interface", typeName)
	}

	for _, method := range interfaceType.Methods.List {
		if len(method.Names) == 0 {
			continue
		}
		if method.Names[0].Name != methodName {
			continue
		}

		signature := exprString(pkg.fset, method.Type)
		for _, fragment := range fragments {
			if !strings.Contains(signature, fragment) {
				t.Fatalf(
					"expected %q.%s signature %q to contain %q",
					typeName,
					methodName,
					signature,
					fragment,
				)
			}
		}
		return
	}

	t.Fatalf("expected %q to have method %q", typeName, methodName)
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

func primaryTagValue(tag string) string {
	if tag == "" {
		return ""
	}
	parts := strings.Split(tag, ",")
	return parts[0]
}

func fieldTagValue(field *ast.Field, key string) string {
	if field == nil || field.Tag == nil {
		return ""
	}
	raw := strings.Trim(field.Tag.Value, "`")
	return reflect.StructTag(raw).Get(key)
}
