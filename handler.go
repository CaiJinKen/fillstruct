package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"io"
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

	resultNode  ast.Node
	isValueSpec bool
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

do:
	h.pkgs = nil // release memory

	if h.f == nil || h.pkg == nil {
		return fmt.Errorf("could not find file %q", h.filepath)
	}

	h.importNames = buildImportNameMap(h.f)

	ast.Inspect(h.f, func(n ast.Node) bool {
		if !h.checkPos(n) {
			return true
		}
		switch n.(type) {
		case *ast.ValueSpec:
			return !h.handValueSpec(n.(*ast.ValueSpec))
		case *ast.FuncDecl:
			return !h.handFuncDecl(n.(*ast.FuncDecl))
		case *ast.ReturnStmt:
			return !h.handReturnStmt(n.(*ast.ReturnStmt))
		default:
			return true
		}
	})

	return
}

// handValueSpec hand ast.ValueSpec
func (h *handler) handValueSpec(node *ast.ValueSpec) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	for i, v := range node.Values {
		litNode, ok := v.(*ast.CompositeLit)
		if !ok {
			continue
		}
		if !h.hintLine(litNode) {
			continue
		}
		node.Values[i] = h.fillCompositeList(litNode)
		isTarget = true
	}

	if h.isValueSpec = isTarget; isTarget {
		h.resultNode = node
	}

	return
}

// handFuncDecl hand ast.FuncDecl
func (h *handler) handFuncDecl(node *ast.FuncDecl) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	for _, v := range node.Body.List {
		if !h.checkPos(v) {
			continue
		}

		switch stmt := v.(type) {
		case *ast.AssignStmt:
			for i, s := range stmt.Rhs {
				switch s.(type) {
				case *ast.FuncLit:
					if isTarget = h.handFuncLit(s.(*ast.FuncLit)); isTarget {
						return
					}
				case *ast.CompositeLit:
					litNode := s.(*ast.CompositeLit)
					if !h.hintLine(litNode) {
						continue
					}
					stmt.Rhs[i] = h.fillCompositeList(litNode)
					isTarget = true
				}
			}
			if isTarget {
				h.resultNode = stmt
				break
			}
		case *ast.DeclStmt:
			h.handDeclStmt(stmt)
		}

	}
	return
}

func (h *handler) handDeclStmt(node *ast.DeclStmt) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	decl, ok := node.Decl.(*ast.GenDecl)
	if !ok {
		return
	}
	for _, v := range decl.Specs {
		if !h.checkPos(v) {
			continue
		}
		switch spec := v.(type) {
		case *ast.ValueSpec:
			isTarget = h.handValueSpec(spec)
			if isTarget {
				goto lable
			}
		}
	}
lable:
	if isTarget {
		h.resultNode = decl
	}
	return
}

// handFuncLit hand ast.FuncLit
func (h *handler) handFuncLit(node *ast.FuncLit) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	for _, v := range node.Body.List {
		if nn, ok := v.(*ast.AssignStmt); ok {
			if isTarget = h.handAssignStmt(nn); isTarget {
				return
			}
		}
	}
	return
}

// handAssignStmt hand ast.AssignStmt
func (h *handler) handAssignStmt(node *ast.AssignStmt) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	for i, s := range node.Rhs {
		litNode, yes := s.(*ast.CompositeLit)
		if !yes {
			continue
		}
		if !h.hintLine(litNode) {
			continue
		}

		node.Rhs[i] = h.fillCompositeList(litNode)

		isTarget = true
		h.resultNode = node
		return
	}
	return
}

// handReturnStmt hand ast.ReturnStmt
func (h *handler) handReturnStmt(node *ast.ReturnStmt) (isTarget bool) {
	if !h.checkPos(node) {
		return
	}
	for _, v := range node.Results {
		litNode, ok := v.(*ast.FuncLit)
		if !ok {
			continue
		}
		if isTarget = h.handFuncLit(litNode); isTarget {
			return
		}
	}
	return
}

// checkPos whether current node contains assigned position
func (h *handler) checkPos(node ast.Node) (ok bool) {
	if node == nil {
		return
	}
	startLine := h.pkg.Fset.Position(node.Pos()).Line
	endLine := h.pkg.Fset.Position(node.End()).Line
	return !(startLine > h.line || endLine < h.line)
}

// hintLine whether the position of the node equals to assigned line
func (h *handler) hintLine(node *ast.CompositeLit) (yes bool) {
	return h.pkg.Fset.Position(node.Rbrace).Line == h.line
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
	var writers []io.Writer
	if *stdOut {
		if *onlyChanged {
			h.printLine()
		} else {
			writers = append(writers, os.Stdout)
		}
	}

	if *writeback {
		f, err := os.OpenFile(h.filepath, os.O_RDWR, 0o66)
		if err != nil {
			return err
		}
		defer f.Close()

		writers = append(writers, f)
	}

	if len(writers) == 0 {
		return
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, h.pkg.Fset, h.f)
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}

	w := io.MultiWriter(writers...)

	_, err = w.Write(data)

	return
}

func (h *handler) printLine() {
	var buf bytes.Buffer
	printer.Fprint(&buf, h.pkg.Fset, h.resultNode)
	data, err := format.Source(buf.Bytes())
	if err != nil {
		return
	}
	// if h.isValueSpec {
	// 	data = append([]byte("var "), data...)
	// }

	os.Stdout.Write(data)
}
