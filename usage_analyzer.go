package main

import (
	"go/ast"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// UsageAnalyzer handles symbol usage detection and call graph analysis
type UsageAnalyzer struct {
	fset        *token.FileSet
	pkgCache    map[string]*packages.Package
	usedSymbols map[string]map[string]bool
}

func NewUsageAnalyzer(fset *token.FileSet, pkgCache map[string]*packages.Package) *UsageAnalyzer {
	return &UsageAnalyzer{
		fset:        fset,
		pkgCache:    pkgCache,
		usedSymbols: make(map[string]map[string]bool),
	}
}

func (ua *UsageAnalyzer) AnalyzeUsage(mainPkg *packages.Package, dependencies []string) error {
	// First, find what symbols are actually referenced from main package
	referencedSymbols := ua.findReferencedSymbols(mainPkg, dependencies)
	
	// Then recursively find all symbols used by those symbols
	for _, depPath := range dependencies {
		ua.usedSymbols[depPath] = make(map[string]bool)
		ua.markUsedSymbols(depPath, referencedSymbols[depPath])
	}
	
	return nil
}

func (ua *UsageAnalyzer) findReferencedSymbols(mainPkg *packages.Package, dependencies []string) map[string]map[string]bool {
	referenced := make(map[string]map[string]bool)
	
	// Initialize maps for all dependencies
	for _, depPath := range dependencies {
		referenced[depPath] = make(map[string]bool)
	}
	
	// Analyze main package files for references to dependency symbols
	for _, file := range mainPkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				// Handle package.Symbol references
				if ident, ok := node.X.(*ast.Ident); ok {
					for _, depPath := range dependencies {
						pkgName := filepath.Base(depPath)
						if ident.Name == pkgName {
							referenced[depPath][node.Sel.Name] = true
						}
					}
				}
			case *ast.CallExpr:
				// Handle constructor calls and type instantiations
				if selExpr, ok := node.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						for _, depPath := range dependencies {
							pkgName := filepath.Base(depPath)
							if ident.Name == pkgName {
								referenced[depPath][selExpr.Sel.Name] = true
							}
						}
					}
				}
			case *ast.CompositeLit:
				// Handle struct literals like &math.Calculator{}
				if selExpr, ok := node.Type.(*ast.SelectorExpr); ok {
					if ident, ok := selExpr.X.(*ast.Ident); ok {
						for _, depPath := range dependencies {
							pkgName := filepath.Base(depPath)
							if ident.Name == pkgName {
								referenced[depPath][selExpr.Sel.Name] = true
							}
						}
					}
				}
			}
			return true
		})
	}
	
	return referenced
}

func (ua *UsageAnalyzer) markUsedSymbols(depPath string, directlyUsed map[string]bool) {
	pkg := ua.pkgCache[depPath]
	if pkg == nil {
		return
	}

	// Start with directly referenced symbols
	toProcess := make(map[string]bool)
	for symbol := range directlyUsed {
		toProcess[symbol] = true
		ua.usedSymbols[depPath][symbol] = true
	}
	
	// Find dependencies of used symbols (methods, embedded types, etc.)
	ua.findSymbolDependencies(depPath, toProcess)
}

func (ua *UsageAnalyzer) findSymbolDependencies(depPath string, symbols map[string]bool) {
	pkg := ua.pkgCache[depPath]
	if pkg == nil {
		return
	}
	
	processed := make(map[string]bool)
	queue := make([]string, 0, len(symbols))
	
	// Initialize queue with directly used symbols
	for symbol := range symbols {
		queue = append(queue, symbol)
	}
	
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		
		if processed[current] {
			continue
		}
		processed[current] = true
		
		// Find what this symbol depends on
		dependencies := ua.getSymbolDependencies(pkg, current)
		for dep := range dependencies {
			if !ua.usedSymbols[depPath][dep] {
				ua.usedSymbols[depPath][dep] = true
				queue = append(queue, dep)
			}
		}
	}
}

func (ua *UsageAnalyzer) getSymbolDependencies(pkg *packages.Package, symbolName string) map[string]bool {
	dependencies := make(map[string]bool)
	
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				// If this is the function we're analyzing
				if node.Name.Name == symbolName {
					// Find what types and functions it uses
					ast.Inspect(node, func(inner ast.Node) bool {
						if ident, ok := inner.(*ast.Ident); ok {
							// Check if this identifier refers to an exported symbol in the same package
							if ident.IsExported() && ident.Name != symbolName {
								if ua.isDefinedInPackage(pkg, ident.Name) {
									dependencies[ident.Name] = true
								}
							}
						}
						return true
					})
				}
			case *ast.TypeSpec:
				// If this is the type we're analyzing
				if node.Name.Name == symbolName {
					// Find embedded types and field types
					ast.Inspect(node.Type, func(inner ast.Node) bool {
						if ident, ok := inner.(*ast.Ident); ok {
							if ident.IsExported() && ident.Name != symbolName {
								if ua.isDefinedInPackage(pkg, ident.Name) {
									dependencies[ident.Name] = true
								}
							}
						}
						return true
					})
				}
			}
			return true
		})
	}
	
	return dependencies
}

func (ua *UsageAnalyzer) isDefinedInPackage(pkg *packages.Package, symbolName string) bool {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.Name == symbolName && d.Name.IsExported() {
					return true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.Name == symbolName && s.Name.IsExported() {
							return true
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.Name == symbolName && name.IsExported() {
								return true
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (ua *UsageAnalyzer) GetUsedSymbols() map[string]map[string]bool {
	return ua.usedSymbols
}