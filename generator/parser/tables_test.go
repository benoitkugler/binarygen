package parser

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	an "github.com/benoitkugler/binarygen/analysis"
	gen "github.com/benoitkugler/binarygen/generator"
)

var ana an.Analyser

func init() {
	os.Remove("../../test-package/source_gen.go")

	var err error
	ana, err = an.NewAnalyser("../../test-package/source_src.go")
	if err != nil {
		panic(err)
	}
}

func TestGenerateParser(t *testing.T) {
	buf := gen.NewBuffer(make(gen.Accu))
	ParsersForFile(ana, &buf)

	content := []byte(fmt.Sprintf(`
	package %s

	// Code generated by binarygen from %s. DO NOT EDIT

	%s
	`, ana.PackageName(), ana.Source, buf.Code(ana.ChildTypes)))

	outfile := "../../test-package/source_gen.go"
	err := os.WriteFile(outfile, content, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	err = exec.Command("goimports", "-w", outfile).Run()
	if err != nil {
		t.Fatal(err)
	}
}
