package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	scopeF = flag.String("scope", ".", "Scope")
	funcF  = flag.String("func", "", "Function or method name")
)

type Func struct {
	Name   string
	Offset int
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}

	fileNames := flag.Args()
	fset, parsedFile, err := parseFiles(fileNames)
	if err != nil {
		log.Fatalf("parse files error: %v", err)
	}

	offsets := getOffsets(fset, parsedFile, *funcF)
	fmt.Println(offsets)

	err = callGuru(*scopeF, fileNames[0], offsets)
	if err != nil {
		log.Fatalf("call guru error: %v", err)
	}
}

func parseFiles(fileNames []string) (*token.FileSet, *ast.File, error) {
	buf := new(bytes.Buffer)
	for _, fileName := range fileNames {
		b, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, nil, err
		}
		buf.Write(b)
	}

	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, "src.go", buf, 0)
	if err != nil {
		return nil, nil, err
	}
	return fset, parsedFile, nil
}

func getOffsets(fset *token.FileSet, parsedFile *ast.File, funcName string) []Func {
	var funcs []Func
	ast.Inspect(parsedFile, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if skipFuncDecl(funcName, x) {
				return true
			}

			funcs = append(funcs, Func{
				Name:   getFuncName(x),
				Offset: fset.Position(x.Name.Pos()).Offset,
			})
			if funcName != "" {
				// we found offset for non empty func name
				return false
			}
		}

		return true
	})
	return funcs
}

func getFuncName(f *ast.FuncDecl) string {
	if f.Recv == nil {
		return f.Name.Name
	}

	expr := f.Recv.List[0].Type
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		expr = starExpr.X
	}

	ident, ok := expr.(*ast.Ident)
	if ok {
		return ident.Name + "." + f.Name.Name
	}

	fmt.Printf("cannot find func name for %s", f.Name.Name)
	return ""
}

func skipFuncDecl(funcName string, f *ast.FuncDecl) bool {
	if funcName == "" {
		return false
	}
	if strings.Contains(funcName, ".") {
		// method
		if f.Recv == nil {
			return true
		}
		if funcName != f.Recv.List[0].Type.(*ast.Ident).Name+"."+f.Name.Name {
			return true
		}
	} else {
		// func
		if f.Recv != nil {
			// skip method
			return true
		}
		if funcName != f.Name.Name {
			return true
		}
	}
	return false
}

func callGuru(scope, fileName string, funcs []Func) error {
	guruPath, err := exec.LookPath("guru")
	if err != nil {
		return err
	}

	for _, _func := range funcs {
		if _func.Name == "main" {
			continue
		}
		offset := fileName + ":#" + fmt.Sprint(_func.Offset)
		args := []string{"-json", "-scope", scope, "referrers", offset}
		//fmt.Println("Running: ", "guru", strings.Join(args, " "))
		cmd := exec.Command(guruPath, args...)
		cmd.Stderr = os.Stderr
		b, err := cmd.Output()
		if err != nil {
			//return err
		}

		fmt.Printf("%s", b)
	}

	return nil
}

func usage() {
	fmt.Printf("pass fileName and funcName\n")
	os.Exit(1)
}
