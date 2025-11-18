package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

type FileBuilder struct {
	// input
	fset       *token.FileSet
	filePkgMap map[string]pkgPath

	// cache
	stdImports map[pkgPath]*ast.ImportSpec
	typeSpecs  []*ast.TypeSpec
	valueSpecs []*ast.ValueSpec
	constDecls []*ast.GenDecl // constはiotaとかあるのでdecl単位
	initDecls  []*ast.FuncDecl
	mainDecl   *ast.FuncDecl
	funcDecls  []*ast.FuncDecl
}

func NewBuilder(fset *token.FileSet, paths map[string]pkgPath) *FileBuilder {
	return &FileBuilder{
		fset:       fset,
		filePkgMap: paths,
		stdImports: make(map[pkgPath]*ast.ImportSpec, 128),
		typeSpecs:  make([]*ast.TypeSpec, 0),
		valueSpecs: make([]*ast.ValueSpec, 0),
		constDecls: make([]*ast.GenDecl, 0),
		initDecls:  make([]*ast.FuncDecl, 0),
		funcDecls:  make([]*ast.FuncDecl, 0),
	}
}

func (b *FileBuilder) commentGroup(t token.Pos) *ast.CommentGroup {
	pos := b.fset.Position(t)
	fp := filepath.ToSlash(pos.Filename)
	name := filepath.Base(fp)
	var pp pkgPath
	var ok bool
	pp, ok = b.filePkgMap[fp]
	if !ok {
		pp = "unknown"
	}

	return &ast.CommentGroup{
		List: []*ast.Comment{
			{Text: fmt.Sprintf("// %s/%s:%d:%d", pp, name, pos.Line, pos.Column)},
		},
	}
}

func (b *FileBuilder) addImportSpec(n *ast.ImportSpec) {
	path := pkgPath(strings.Trim(n.Path.Value, `"`))
	if isStd(path) {
		b.stdImports[path] = n
	}
}

func (b *FileBuilder) addTypeSpec(n *ast.TypeSpec) {
	b.typeSpecs = append(b.typeSpecs, n)
}

func (b *FileBuilder) addValueSpec(n *ast.ValueSpec) {
	b.valueSpecs = append(b.valueSpecs, n)
}

func (b *FileBuilder) addConstDecl(n *ast.GenDecl) {
	if n.Tok == token.CONST {
		b.constDecls = append(b.constDecls, n)
	}
}

func (b *FileBuilder) addInitDecl(n *ast.FuncDecl) {
	b.initDecls = append(b.initDecls, n)
}

func (b *FileBuilder) setMainDecl(n *ast.FuncDecl) {
	b.mainDecl = n
}

func (b *FileBuilder) addFuncDecl(n *ast.FuncDecl) {
	b.funcDecls = append(b.funcDecls, n)
}

func (b *FileBuilder) Build() (*ast.File, error) {
	// check required values
	if b.mainDecl == nil {
		return nil, errors.New("main function not found")
	}

	file := &ast.File{
		Name: ast.NewIdent("main"),
	}

	// add imports
	if len(b.stdImports) > 0 {
		importDecl := &ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: make([]ast.Spec, 0, len(b.stdImports)),
		}
		paths := make([]string, 0, len(b.stdImports))
		for p := range b.stdImports {
			paths = append(paths, string(p))
		}
		sort.Strings(paths)
		for _, p := range paths {
			importDecl.Specs = append(importDecl.Specs, b.stdImports[pkgPath(p)])
		}
		file.Decls = append(file.Decls, importDecl)

	}

	// add inits
	if len(b.initDecls) > 0 {
		initDecl := &ast.FuncDecl{
			Name: ast.NewIdent("init"),
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{},
			},
		}
		file.Decls = append(file.Decls, initDecl)
		for _, d := range b.initDecls {
			stmt := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: d.Name,
				},
			}
			initDecl.Body.List = append(initDecl.Body.List, stmt)
			file.Decls = append(file.Decls, d)
		}
	}

	// add types
	for _, v := range b.typeSpecs {
		decl := &ast.GenDecl{
			Tok:   token.TYPE,
			Specs: []ast.Spec{v},
			Doc:   b.commentGroup(v.Pos()),
		}
		file.Decls = append(file.Decls, decl)
	}

	// add values
	for _, v := range b.valueSpecs {
		decl := &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{v},
			Doc:   b.commentGroup(v.Pos()),
		}
		file.Decls = append(file.Decls, decl)
	}

	// add consts
	for _, d := range b.constDecls {
		d.Doc = b.commentGroup(d.Pos())
		file.Decls = append(file.Decls, d)
	}

	// add funcs
	mainDecl := &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: b.mainDecl.Body,
		Doc:  b.commentGroup(b.mainDecl.Pos()),
	}
	file.Decls = append(file.Decls, mainDecl)
	for _, d := range b.funcDecls {
		d.Doc = b.commentGroup(d.Pos())
		file.Decls = append(file.Decls, d)
	}

	return file, nil
}
