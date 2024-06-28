package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// handler include file check, travel ast file, fill struct with type-zero value, write back
type handler struct {
	// param
	filepath string
	line     int
	// internal use
	pkgs        []*packages.Package
	pkg         *packages.Package
	f           *ast.File
	importNames map[string]string // import path -> import name
}

func newHandler(filepath string, line int) *handler {
	return &handler{
		filepath:    filepath,
		line:        line,
		pkg:         nil,
		importNames: nil,
	}
}

// preCheck check file status & build packages
func (h *handler) preCheck() (err error) {
	h.line = *line

	path, err := absPath(*filename)
	if err != nil {
		return
	}
	h.filepath = path

	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: true,
		Dir:   filepath.Dir(path),
		Fset:  token.NewFileSet(),
		Env:   os.Environ(),
	})
	if err != nil {
		return
	}
	h.pkgs = pkgs
	return
}

// travel packages to find the assigned file to hand ast.Node
func (h *handler) travel() (err error) {
	for _, p := range h.pkgs {
		for _, af := range p.Syntax {
			if file := p.Fset.File(af.Pos()); file.Name() != h.filepath {
				continue
			}

			h.f = af
			h.pkg = p
			goto do
		}
	}
	h.pkgs = nil // release memory

do:

	if h.f == nil || h.pkg == nil {
		return fmt.Errorf("could not find file %q", h.filepath)
	}

	h.importNames = buildImportNameMap(h.f)

	ast.Inspect(h.f, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		startLine := h.pkg.Fset.Position(n.Pos()).Line
		endLine := h.pkg.Fset.Position(n.End()).Line

		if startLine > h.line || endLine < h.line {
			return true
		}
		switch n.(type) {
		case *ast.ValueSpec:
			return !h.handValueSpec(n.(*ast.ValueSpec))
		case *ast.FuncDecl:
			return !h.handFuncDecl(n.(*ast.FuncDecl))
		default:
			return true
		}

	})

	return

}

// handValueSpec hand ast.ValueSpec
func (h *handler) handValueSpec(node *ast.ValueSpec) (result bool) {
	for i, v := range node.Values {
		lll, ok := v.(*ast.CompositeLit)
		if !ok {
			continue
		}
		node.Values[i] = h.fillCompositeList(lll)
		result = true
	}

	return
}

// handFuncDecl hand ast.FuncDecl
func (h *handler) handFuncDecl(node *ast.FuncDecl) (result bool) {
	for _, v := range node.Body.List {
		stmt, ok := v.(*ast.AssignStmt)
		if !ok {
			continue
		}
		if h.pkg.Fset.Position(stmt.TokPos).Line != h.line {
			continue
		}
		for i, s := range stmt.Rhs {
			lll, yes := s.(*ast.CompositeLit)
			if !yes {
				continue
			}
			stmt.Rhs[i] = h.fillCompositeList(lll)

			result = true
		}

	}
	return
}

// fillCompositeList gen assigned zero value
func (h *handler) fillCompositeList(node *ast.CompositeLit) (result ast.Expr) {
	var info litInfo
	var prev types.Type
	var ok bool
	info.name, _ = h.pkg.TypesInfo.Types[node].Type.(*types.Named)
	info.typ, ok = h.pkg.TypesInfo.Types[node].Type.Underlying().(*types.Struct)
	if !ok {
		prev = h.pkg.TypesInfo.Types[node].Type.Underlying()
		return
	}
	info.hideType = hideType(prev)
	result, _ = zeroValue(h.pkg.Types, h.importNames, node, info)
	return
}

// writeBack write back to the source file
func (h *handler) writeBack() (err error) {
	var buf bytes.Buffer
	printer.Fprint(&buf, h.pkg.Fset, h.f)
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}

	w, err := os.OpenFile(h.filepath, os.O_RDWR, 066)
	if err != nil {
		return
	}
	defer w.Close()

	_, err = w.Write(data)

	return nil

}
