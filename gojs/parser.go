package gojs

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"

	"github.com/gogap/gocoder"
)

//go:embed templates/*
var tmplData embed.FS

type GenerateOptions struct {
	TemplateName string
	PackagePath  string
	PackageAlias string
	GoPath       string
	ProjectPath  string

	TemplateVars *TemplateVars

	Args interface{}
}

func Parser(pkgPath string, gopath string) (vars *TemplateVars, err error) {
	pkg, err := gocoder.NewGoPackage(pkgPath,
		gocoder.OptionGoPath(gopath),
	)

	if err != nil {
		return
	}

	tmplVars := &TemplateVars{
		PackageName:  pkg.Name(),
		PackagePath:  pkg.Path(),
		PackageFuncs: make(map[string]string),
		PackageTypes: make(map[string]string),
		PackageVars:  make(map[string]string),
	}

	numFuncs := pkg.NumFuncs()

	for i := 0; i < numFuncs; i++ {
		goFunc := pkg.Func(i)

		if !isExported(goFunc.Name()) {
			continue
		}

		if len(goFunc.Receiver()) > 0 {
			continue
		}

		// 首字母小写
		funcName := strings.ToLower(string(goFunc.Name()[0])) + goFunc.Name()[1:]
		tmplVars.PackageFuncs[funcName] = goFunc.Name()
	}

	numVars := pkg.NumVars()

	for i := 0; i < numVars; i++ {
		goVar := pkg.Var(i)

		if !isExported(goVar.Name()) {
			continue
		}

		tmplVars.PackageVars[goVar.Name()] = goVar.Name()
	}

	numTypes := pkg.NumTypes()

	for i := 0; i < numTypes; i++ {
		goType := pkg.Type(i)

		if !isExported(goType.Name()) {
			continue
		}

		if goType.IsStruct() {
			tmplVars.PackageTypes[goType.Name()] = goType.Name()
		}

	}

	vars = tmplVars

	return
}

func GenerateCode(options GenerateOptions) (code string, err error) {

	tmplBytes, err := tmplData.ReadFile(fmt.Sprintf("templates/%s.tmpl", options.TemplateName))
	if err != nil {
		return
	}

	tmpl, err := template.New(options.TemplateName).Funcs(templateFuncs()).Parse(string(tmplBytes))
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(nil)

	err = tmpl.Execute(buf, options.TemplateVars)

	if err != nil {
		return
	}

	if len(options.PackageAlias) > 0 {
		options.TemplateVars.PackageName = options.PackageAlias
	}

	codeBytes, err := format.Source(buf.Bytes())

	if err != nil {
		return
	}

	code = string(codeBytes)

	return
}

func isExported(v string) bool {
	if len(v) == 0 {
		return false
	}

	if v[0] >= 'A' && v[0] <= 'Z' {
		return true
	}

	return false
}

func templateFuncs() map[string]interface{} {
	return map[string]interface{}{
		"exist": func(v map[string]string, key string) bool {
			_, exist := v[key]
			return exist
		},
		"toTitle": func(v string) string {
			if len(v) == 0 {
				return v
			}

			return strings.ToUpper(string(v[0])) + v[1:]
		},
	}
}
