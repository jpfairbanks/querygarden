package main

import (
	"fmt"
	"os"
	"testing"
	"text/template"
	"github.com/jpfairbanks/featex/querygen/depr"
	//"log"
)

func TestProcessTemplateFile(t *testing.T) {
	var err error
	args := []string{"templates/demographics.struct.sql.tmpl"}

	t.Log("ARGS:%s", args)
	os.Stdout.Sync()
	p := Params{"table",
		"year <= 2016 and year >= 2000",
		"limit 100"}

	filen := args[0]
	outfile := os.Stdout
	if len(args) > 1 {
		t.Log("Writing Rendered Template to file:%s", args[1])
		outfile, err = os.Create(args[1])
		if err != nil {
			t.Fatalf("Cannot open output file supplied: %v", err)
		}
	}
	err = depr.ProcessTemplate(outfile, filen, p)
	if err != nil {
		t.Fatalf("Failed to ProcessTemplate:%s:%s", filen, err)
	}
}

func TestProcessTemplate(t *testing.T) {
	args := os.Args[1:]
	fmt.Println(args)
	p := Params{"table",
		"year <= 2016 and year >= 2000",
		"limit 100"}

	templ := template.New("test_string")
	s := "select gender, race, age, count(*) from {{.Table}} where {{.Cond}} group by gender, race, age {{.Limit}}\n"
	templ, err := templ.Parse(s)
	if err != nil {
		t.Log(t)
		t.Fatal(err)
	}

	// log.Printf("Template t: %v\n", t)
	// log.Printf("t.Tree: %v\n", t.Tree)
	err = templ.Execute(os.Stdout, p)
	if err != nil {
		t.Log("Failed to execute")
		t.Fatal(err)
	}
}
