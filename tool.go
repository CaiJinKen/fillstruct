package main

import (
	"go/ast"
	"go/types"
	"path/filepath"
)

// absPath returns the full path of the filename
func absPath(filename string) (string, error) {
	eval, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return "", err
	}
	return filepath.Abs(eval)
}

// hideType returns true when t is array || map || slice
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

// buildImportNameMap get all imported packages in file
func buildImportNameMap(f *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, i := range f.Imports {
		if i.Name != nil && i.Name.Name != "_" {
			path := i.Path.Value
			imports[path[1:len(path)-1]] = i.Name.Name
		}
		if i.Name == nil {
			path := i.Path.Value
			path = path[1 : len(path)-1]
			_, name := filepath.Split(path)
			if name == "." {
				continue
			}
			imports[path] = name
		}
	}
	return imports
}
