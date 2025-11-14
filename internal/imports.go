package internal

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ImportSet stores unique imports using a map.
type ImportSet map[string]bool
type OutputPackage string
type IncludeModelAlias bool

// New returns a new ImportSet.
func New(structName, output string) (OutputPackage, ImportSet, IncludeModelAlias) {
	set := ImportSet(make(map[string]bool))
	set.Add("fmt")
	set.Add("context")
	set.Add("database/sql")
	modelPkgPath, err := findModelPkg(structName)
	if err != nil {
		panic(err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule,
	}
	pkgs, err := packages.Load(cfg, output)
	if err != nil {
		panic(err)
	}
	outPkgPath := pkgs[0].PkgPath
	includePkgModel := false
	if modelPkgPath != outPkgPath {
		set.Add(fmt.Sprintf(`dto "%v"`, modelPkgPath))
		includePkgModel = true
	}
	return OutputPackage(pkgs[0].Name), set, IncludeModelAlias(includePkgModel)

}

// Add inserts an import path.
func (s ImportSet) Add(path string) {
	if strings.TrimSpace(path) != "" {
		s[path] = true
	}
}

// Sorted returns a sorted slice of imports.
func (s ImportSet) Sorted() []string {
	result := make([]string, 0, len(s))
	for p := range s {
		result = append(result, p)
	}
	sort.Strings(result)
	return result
}

// SplitStdAndExternal separates stdlib imports from others.
func (s ImportSet) splitStdAndExternal() (stdlib []string, external []string) {
	for _, p := range s.Sorted() {
		if isStdPackage(p) {
			stdlib = append(stdlib, p)
		} else {
			external = append(external, p)
		}
	}
	return
}

// Detect if an import belongs to the Go standard library.
func isStdPackage(path string) bool {
	// If import path contains a dot, it's not standard library.
	if strings.Contains(path, ".") {
		return false
	}

	// build.Import with empty srcDir identifies stdlib
	_, err := build.Default.Import(path, "", build.IgnoreVendor)
	return err == nil
}

type ImportGroups struct {
	Std []string
	Ext []string
}

func (s ImportSet) Group() ImportGroups {
	std, ext := s.splitStdAndExternal() // internal helper
	return ImportGroups{Std: std, Ext: ext}
}

func findModelPkg(structName string) (pkgPath string, err error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedModule | packages.NeedFiles | packages.NeedImports,
		Tests: true,
	}

	// Load all packages under current module
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return "", err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if ts.Name.Name == structName {
						// Found the struct!
						return pkg.String(), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("struct %s not found", structName)
}
