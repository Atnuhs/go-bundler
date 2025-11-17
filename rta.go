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
		declRoots:   make(map[types.Object]bool, 128),
	}
	a.buildSSA()
	a.analyzeRTA()
	a.buildNodeToFn()
	a.collectReferencedDecls()
	a.buildDeclGraph()
	a.propagateDeclReachability()
	a.closeMethodsByType()
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
	nodeToFn    map[ast.Node]*ssa.Function
	declGraph   map[types.Object][]types.Object
	declRoots   map[types.Object]bool

	// output
	reachableDecls map[types.Object]bool
}

func (a *ReachabilityAnalyzer) buildSSA() {
	prog, ssaPkgs := ssautil.AllPackages([]*packages.Package{a.mainPkg}, ssa.InstantiateGenerics)
	prog.Build()

	// cache
	a.prog = prog
	a.ssaPkgs = ssaPkgs
}

func (a *ReachabilityAnalyzer) analyzeRTA() {
	roots := rootsPkgs(a.ssaPkgs)
	res := rta.Analyze(roots, true)
	for fn := range res.Reachable {
		a.reachableFn[fn] = true
	}
}

func (a *ReachabilityAnalyzer) buildNodeToFn() {
	a.nodeToFn = make(map[ast.Node]*ssa.Function, 128)
	for _, ssaPkg := range a.ssaPkgs {
		for _, mem := range ssaPkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				if syn := fn.Syntax(); syn != nil {
					a.nodeToFn[syn] = fn
				}
			}
		}
	}
}

func (a *ReachabilityAnalyzer) collectReferencedDecls() {
	for _, p := range a.topoPkgs {
		info := p.TypesInfo
		if info == nil {
			continue
		}
		for _, f := range p.Syntax {
			v := &declUseVisitor{
				info:          info,
				nodeToFn:      a.nodeToFn,
				reachableFunc: a.reachableFn,
				reachableDecl: a.declRoots,
			}
			ast.Walk(v, f)
		}
	}
}

func (a *ReachabilityAnalyzer) buildDeclGraph() {
	a.declGraph = make(map[types.Object][]types.Object, 128)

	for _, p := range a.topoPkgs {
		info := p.TypesInfo
		if info == nil {
			continue
		}

		// A, B = Cならば、A -> C, B -> Cの辺を張りたい
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
	a.reachableDecls = make(map[types.Object]bool, len(a.declRoots))
	// seed reachable decls
	queue := make([]types.Object, 0, len(a.declRoots))
	for f := range a.reachableFn {
		if obj := f.Object(); obj != nil {
			if f.Pkg != nil && !isStd(pkgPath(f.Pkg.Pkg.Path())) {
				queue = append(queue, obj)
			}
		}
	}

	// bfs
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:] // pop
		if a.reachableDecls[cur] {
			continue
		}
		a.reachableDecls[cur] = true

		for _, next := range a.declGraph[cur] {
			if !a.reachableDecls[next] {
				queue = append(queue, next) // push
			}
		}
	}
}

func (a *ReachabilityAnalyzer) closeMethodsByType() {
	for _, pkg := range a.topoPkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok || !a.reachableDecls[tn] {
				continue
			}

			// 利用のある型のメソッドを全部reachableにする
			// interface越しの利用を検知できないので。
			T := tn.Type()
			for _, typ := range []types.Type{T, types.NewPointer(T)} {
				mset := types.NewMethodSet(typ)
				for i := 0; i < mset.Len(); i++ {
					sel := mset.At(i)
					if fn, ok := sel.Obj().(*types.Func); ok {
						a.reachableDecls[fn] = true
					}
				}
			}
		}
	}
}

func rootsPkgs(pkgs []*ssa.Package) []*ssa.Function {
	roots := make([]*ssa.Function, 0, 128)
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

type declUseVisitor struct {
	info          *types.Info
	nodeToFn      map[ast.Node]*ssa.Function
	reachableFunc map[*ssa.Function]bool
	reachableDecl map[types.Object]bool

	stack   []ast.Node
	funcIds []int
}

func (v *declUseVisitor) Visit(n ast.Node) ast.Visitor {
	// nがnilなのはそのサブツリーを抜ける時
	if n == nil {
		if len(v.stack) == 0 {
			return v
		}
		v.stack = v.stack[:len(v.stack)-1]
		if len(v.funcIds) > 0 && len(v.stack) == v.funcIds[len(v.funcIds)-1] {
			v.funcIds = v.funcIds[:len(v.funcIds)-1]
		}
		return v
	}

	// nodeに入るタイミング
	switch n.(type) {
	case *ast.FuncDecl, *ast.FuncLit:
		if _, ok := v.nodeToFn[n]; ok {
			v.funcIds = append(v.funcIds, len(v.stack))
		}
	}
	v.stack = append(v.stack, n)

	var cur *ssa.Function
	if len(v.funcIds) > 0 {
		lastFuncNode := v.stack[v.funcIds[len(v.funcIds)-1]]
		cur = v.nodeToFn[lastFuncNode]
	}

	if id, ok := n.(*ast.Ident); ok {
		if obj := v.info.Uses[id]; obj != nil {
			if cur != nil && v.reachableFunc[cur] {
				v.reachableDecl[obj] = true
			}
		}
	}
	return v
}
