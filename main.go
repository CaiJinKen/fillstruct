package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

var errNotFound = errors.New("no struct literal found at selection")

var File string

func main() {
	log.SetFlags(0)
	log.SetPrefix("fillstruct: ")

	var (
		filename = flag.String("file", "", "filename")
		line     = flag.Int("line", 0, "line number of the struct literal, optional if -offset is present")
	)
	flag.Parse()

	if *line == 0 || *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	path, err := absPath(*filename)
	if err != nil {
		log.Fatal(err)
	}
	File = path

	var overlay map[string][]byte

	cfg := &packages.Config{
		Overlay: overlay,
		Mode:    packages.LoadAllSyntax,
		Tests:   true,
		Dir:     filepath.Dir(path),
		Fset:    token.NewFileSet(),
		Env:     os.Environ(),
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if *line > 0 {
		err = byLine(pkgs, path, *line)
		switch err {
		case nil:
			return
		default:
			log.Fatal(err)
		}
	}

	log.Fatal(errNotFound)
}

func absPath(filename string) (string, error) {
	eval, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return "", err
	}
	return filepath.Abs(eval)
}

func byLine(lprog []*packages.Package, path string, line int) (err error) {
	var f *ast.File
	var pkg *packages.Package
	for _, p := range lprog {
		for _, af := range p.Syntax {
			if file := p.Fset.File(af.Pos()); file.Name() == path {
				f = af
				pkg = p
			}
		}
	}
	if f == nil || pkg == nil {
		return fmt.Errorf("could not find file %q", path)
	}
	importNames := buildImportNameMap(f)

	var prev types.Type
	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		startLine := pkg.Fset.Position(lit.Pos()).Line
		endLine := pkg.Fset.Position(lit.End()).Line

		if !(startLine <= line && line <= endLine) {
			return true
		}
		rawValues := lit.Values
		lit.Values = nil

		for _, v := range rawValues {
			lll, ok := v.(*ast.CompositeLit)
			if !ok {
				continue
			}

			var info litInfo
			info.name, _ = pkg.TypesInfo.Types[lll].Type.(*types.Named)
			info.typ, ok = pkg.TypesInfo.Types[lll].Type.Underlying().(*types.Struct)
			if !ok {
				prev = pkg.TypesInfo.Types[lll].Type.Underlying()
				err = errNotFound
				return true
			}
			info.hideType = hideType(prev)
			newlit, _ := zeroValue(pkg.Types, importNames, lll, info)
			lit.Values = append(lit.Values, newlit)

		}

		return false
	})
	var buf bytes.Buffer
	printer.Fprint(&buf, pkg.Fset, f)

	w, err := os.OpenFile(File, os.O_RDWR, 066)
	w.Write(buf.Bytes())
	defer w.Close()

	if err != nil {
		return err
	}

	return nil

}

func hideType(t types.Type) bool {
	switch t.(type) {
	case *types.Array:
		return true
	case *types.Map:
		return true
	case *types.Slice:
		return true
	default:
		return false
	}
}

func buildImportNameMap(f *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, i := range f.Imports {
		if i.Name != nil && i.Name.Name != "_" {
			path := i.Path.Value
			imports[path[1:len(path)-1]] = i.Name.Name
		}
	}
	return imports
}
