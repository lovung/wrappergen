package internal

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Method struct {
	Name   string
	Params string
	Args   string
	Return string
}

type Interface struct {
	Name    string
	Methods []*Method

	parrentPackage *packages.Package
	typeInterface  *types.Interface
	astTypeSpec    *ast.TypeSpec
}

type WrapperGenerator struct {
	packagePath       string
	excludeFiles      map[string]bool
	excludeInterfaces map[string]bool

	foundPackage   *packages.Package
	foundInterface []*Interface
}

type Option func(*WrapperGenerator)

func WithIgnoreFileNames(fileNames []string) Option {
	return func(g *WrapperGenerator) {
		if g.excludeFiles == nil {
			g.excludeFiles = make(map[string]bool)
		}
		for i := range fileNames {
			g.excludeFiles[fileNames[i]] = true
		}
	}
}

func WithIgnoreInterfaceNames(fileNames []string) Option {
	return func(g *WrapperGenerator) {
		if g.excludeInterfaces == nil {
			g.excludeInterfaces = make(map[string]bool)
		}
		for i := range fileNames {
			g.excludeInterfaces[fileNames[i]] = true
		}
	}
}

func NewWrapperGenerator(packagePath string, options ...Option) WrapperGenerator {
	g := WrapperGenerator{
		packagePath: packagePath,
	}
	for _, o := range options {
		o(&g)
	}
	return g
}

func (g *WrapperGenerator) ParseData() (string, []*Interface, error) {
	err := g.getPackage()
	if err != nil {
		return "", nil, err
	}
	interfaceList, err := g.loadAllInterfaces()
	if err != nil {
		return "", nil, err
	}

	for _, iface := range interfaceList {
		methods, err := g.parseMethods(iface)
		if err != nil {
			return "", nil, err
		}
		iface.Methods = methods
	}

	g.foundInterface = interfaceList
	return g.foundPackage.Name, g.foundInterface, nil
}

func (g *WrapperGenerator) parseMethods(iface *Interface) ([]*Method, error) {
	methods := make([]*Method, 0, iface.typeInterface.NumMethods())
	for i := 0; i < iface.typeInterface.NumMethods(); i++ {
		method := iface.typeInterface.Method(i)

		params, args, err := g.parseMethodParams(method)
		if err != nil {
			return nil, err
		}

		returnType, err := g.parseReturnType(method)
		if err != nil {
			return nil, err
		}
		methods = append(methods, &Method{
			Name:   method.Name(),
			Params: params,
			Args:   args,
			Return: returnType,
		})
	}
	return methods, nil
}

func (g *WrapperGenerator) parseTypes(tuple *types.Tuple, isVariadic, withName, withType bool) []string {
	var typeStrings []string

	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		typeStr := types.TypeString(param.Type(), func(p *types.Package) string {
			if p.Name() == g.foundPackage.Name {
				return ""
			}
			return afterLastSplash(p.Name())
		})
		var str string
		if withName {
			if isVariadic && i == tuple.Len()-1 {
				str += strings.ReplaceAll(param.Name(), "[]", "")
				if !withType {
					str += "..."
				}
			} else {
				str += param.Name()
			}
		}
		if withType {
			if isVariadic && i == tuple.Len()-1 {
				str += " " + strings.ReplaceAll(typeStr, "[]", "...")
			} else {
				str += " " + typeStr
			}
		}
		typeStrings = append(typeStrings, str)
	}

	return typeStrings
}

func (g *WrapperGenerator) parseMethodParams(method *types.Func) (string, string, error) {
	var params []string
	var args []string

	sig, ok := method.Type().(*types.Signature)
	if !ok {
		return "", "", fmt.Errorf("invalid method type")
	}

	paramsList := sig.Params()
	params = g.parseTypes(paramsList, sig.Variadic(), true, true)
	args = g.parseTypes(paramsList, sig.Variadic(), true, false)

	return strings.Join(params, ", "), strings.Join(args, ", "), nil
}

func (g *WrapperGenerator) parseReturnType(method *types.Func) (string, error) {
	sig, ok := method.Type().(*types.Signature)
	if !ok {
		return "", fmt.Errorf("invalid method type")
	}

	tuple := sig.Results()
	returns := g.parseTypes(tuple, sig.Variadic(), false, true)

	return strings.Join(returns, ", "), nil
}

func (g *WrapperGenerator) getPackage() error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, g.packagePath)
	if err != nil {
		return err
	}

	if packages.PrintErrors(pkgs) > 0 || len(pkgs) == 0 {
		return fmt.Errorf("failed to load package %s", g.packagePath)
	}

	g.foundPackage = pkgs[0]
	return nil
}

func (g *WrapperGenerator) loadAllInterfaces() ([]*Interface, error) {
	ifaces := make([]*Interface, 0)
	// Search all files
	for _, syntax := range g.foundPackage.Syntax {
		fileName := afterLastSplash(g.foundPackage.Fset.Position(syntax.Pos()).Filename)
		if g.excludeFiles[fileName] {
			continue
		}
		for _, decl := range syntax.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						obj := g.foundPackage.TypesInfo.Defs[typeSpec.Name]
						if obj == nil {
							continue
						}
						if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
							if g.excludeInterfaces[typeSpec.Name.String()] {
								continue
							}
							ifaces = append(ifaces, &Interface{
								Name:           typeSpec.Name.String(),
								parrentPackage: g.foundPackage,
								typeInterface:  iface,
								astTypeSpec:    typeSpec,
							})
						}
					}
				}
			}
		}
	}

	return ifaces, nil
}

func afterLastSplash(typeStr string) string {
	lastIndex := strings.LastIndex(typeStr, "/")
	if lastIndex == -1 {
		return typeStr
	}
	return typeStr[lastIndex+1:]
}
