package main

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func AnalyzeReachableDecls(main *packages.Package, topoPkg []*packages.Package) map[types.Object]bool {
	a := &ReachabilityAnalyzer{
		mainPkg:     main,
		topoPkgs:    topoPkg,
		reachableFn: make(map[*ssa.Function]bool, 128),
	}
	a.buildSSA()
	a.analyzeRTA()
	a.buildDeclGraph()
	a.propagateDeclReachability()
	return a.reachableDecls
}

type ReachabilityAnalyzer struct {
	// input
	mainPkg  *packages.Package
	topoPkgs []*packages.Package

	// cache
	prog        *ssa.Program
	ssaPkgs     []*ssa.Package
	reachableFn map[*ssa.Function]bool
	declGraph   map[types.Object][]types.Object

	// output
	reachableDecls map[types.Object]bool
}

func (a *ReachabilityAnalyzer) buildSSA() {
	prog, ssaPkgs := ssautil.AllPackages([]*packages.Package{a.mainPkg}, ssa.InstantiateGenerics)
	prog.Build()

	a.prog = prog
	a.ssaPkgs = ssaPkgs
}

func (a *ReachabilityAnalyzer) analyzeRTA() {
	roots := rootsPkgs(a.ssaPkgs)
	res := rta.Analyze(roots, true)
	if res == nil {
		panic("res is nil")
	}
	for fn := range res.Reachable {
		a.reachableFn[fn] = true
	}
}

func (a *ReachabilityAnalyzer) buildDeclGraph() {
	a.declGraph = make(map[types.Object][]types.Object, 128)

	for _, p := range a.topoPkgs {
		info := p.TypesInfo
		if info == nil {
			continue
		}

		for _, f := range p.Syntax {
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					obj := info.Defs[d.Name]
					if obj != nil {
						a.inspectDeclBody(info, []types.Object{obj}, d)
					}
				case *ast.GenDecl:
					switch d.Tok {
					case token.TYPE:
						for _, spec := range d.Specs {
							if ts, ok := spec.(*ast.TypeSpec); ok {
								if obj := info.Defs[ts.Name]; obj != nil {
									a.inspectDeclBody(info, []types.Object{obj}, d)
								}
							}
						}
					case token.VAR, token.CONST:
						for _, spec := range d.Specs {
							if vs, ok := spec.(*ast.ValueSpec); ok {
								var curDecls []types.Object
								for _, name := range vs.Names {
									if obj := info.Defs[name]; obj != nil {
										curDecls = append(curDecls, obj)
									}
								}
								if len(curDecls) > 0 {
									a.inspectDeclBody(info, curDecls, vs)
								}
							}
						}
					}
				}
			}
		}
	}
}

func (a *ReachabilityAnalyzer) inspectDeclBody(info *types.Info, parents []types.Object, root ast.Node) {
	ast.Inspect(root, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if obj := info.Uses[id]; obj != nil {
				for _, p := range parents {
					a.declGraph[p] = append(a.declGraph[p], obj)
				}
			}
		}
		return true
	})
}

func (a *ReachabilityAnalyzer) propagateDeclReachability() {
	a.reachableDecls = make(map[types.Object]bool, len(a.reachableFn))
	queue := make([]types.Object, 0, len(a.reachableFn))
	for f := range a.reachableFn {
		if obj := f.Object(); obj != nil {
			if f.Pkg != nil && !isStd(pkgPath(f.Pkg.Pkg.Path())) {
				queue = append(queue, obj)
			}
		}
	}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if a.reachableDecls[cur] {
			continue
		}
		a.reachableDecls[cur] = true

		if tn, ok := cur.(*types.TypeName); ok {
			for _, m := range methodOfType(tn) {
				if !a.reachableDecls[m] {
					queue = append(queue, m)
				}
			}
		}

		for _, next := range a.declGraph[cur] {
			if !a.reachableDecls[next] {
				queue = append(queue, next)
			}
		}
	}
}

func methodOfType(tn *types.TypeName) []types.Object {
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return nil
	}
	if named.Obj().Pkg() == nil {
		return nil
	}
	ret := make([]types.Object, 0, named.NumMethods())
	for i := 0; i < named.NumMethods(); i++ {
		ret = append(ret, named.Method(i))
	}
	return ret
}

func rootsPkgs(pkgs []*ssa.Package) []*ssa.Function {
	roots := make([]*ssa.Function, 0, 2)
	for _, p := range pkgs {
		if p == nil || p.Pkg.Name() != "main" {
			continue
		}
		if f := p.Func("init"); f != nil {
			roots = append(roots, f)
		}
		if f := p.Func("main"); f != nil {
			roots = append(roots, f)
		}
	}
	return roots
}
