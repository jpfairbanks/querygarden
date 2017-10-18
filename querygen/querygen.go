package main

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"
)

type Params struct {
	Table string
	Cond  string
	Limit string
}

// A Context hold the information for where to find templates and put rendered results
type Context struct {
	templatedir string
	outputdir   string
}

// A Scope holds the name of the currently executing scope along with a Reader containing the template,
// a writer to hold the result,
// and the parameters that is a map[string]interface{} full of variable bindings.
type Scope struct {
	scopename  string
	r          io.Reader
	w          io.Writer
	parameters map[string]interface{}
}

// FilesFromConfig opens the template for reading and the output for writing
// filenames are taken from the config file
// Template: input file
// Filename: output file
func (c Context) FilesFromConfig(scopename string) (io.Reader, io.Writer, error) {
	var filen string
	var outfile io.Writer
	var tfile io.Reader
	var err error = nil

	// input file
	filen = viper.GetString(fmt.Sprintf("%s.Template", scopename))
	filen = path.Join(c.templatedir, filen)
	tfile, err = os.Open(filen)
	if err != nil {
		return tfile, outfile, err
	}
	// output file
	if c.outputdir == "-" {
		outfile = os.Stdout
	} else {
		err = os.MkdirAll(c.outputdir, 0755)
		if err != nil {
			log.Printf("Cannot make directory for output files:%s", c.outputdir)
		}
		s := viper.GetString(fmt.Sprintf("%s.Filename", scopename))
		s = path.Join(c.outputdir, s)
		outfile, err = os.Create(s)
		if err != nil {
			log.Printf("Cannot open output file as specified in config:%s", s)
			return tfile, outfile, err
		}
	}
	return tfile, outfile, nil
}

// ProcessTemplate takes a scope that is ready to go and executes it.
func (sp *Scope) ProcessTemplate() error {
	t := template.New(sp.scopename)
	s, err := ReadString(sp.r)
	if err != nil {
		log.Printf("%s:Could not read file %s\n", sp.scopename, sp.r)
		return err
	}
	t, err = t.Parse(s)
	if err != nil {
		log.Println("Failed to parse: %s\n", err)
		return err
	}
	if sp.w == nil {
		return errors.New("Output io.Writer was nil")
	}
	if sp.parameters == nil{
		errors.New("scope.parameters were not initialized you must call Params() before ProcessTemplate.")
	}
	err = t.Execute(sp.w, sp.parameters)
	if err != nil {
		log.Printf("Failed to execute: %s\n", err)
	}
	return err
}

// Params loads the variable bindings from the config systems into the scope and returns it.
// This method must be called before ProcessTemplate()
func (sp *Scope) Params() (map[string]interface{}, error) {
	globalparams := viper.GetStringMap("global")
	localparams := viper.GetStringMap(sp.scopename)
	scope := make(map[string]interface{})
	scope = MergeMaps(scope, globalparams)
	scope = MergeMaps(scope, localparams)
	log.Printf("global params from config:%s\n", globalparams)
	log.Printf("local params from config:%s\n", localparams)
	log.Printf("actual params from config:%s\n", scope)
	sp.parameters = scope
	return scope, nil
}

func config() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/querygen/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.querygen") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	defglobal := map[string]interface{}{
		"schema":  "schema",
		"version": "1.0",
	}
	viper.SetDefault("global", defglobal)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}

func ReadString(file io.Reader) (string, error) {
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func MergeMaps(dst map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func Execute(scopename string) error {
	var tfile io.Reader
	var outfile io.Writer
	var err error

	ctx := Context{"templates", "sql"}
	tfile, outfile, err = ctx.FilesFromConfig(scopename)
	if err != nil {
		log.Printf("Skipping scope because of bad files: %s", scopename)
		log.Printf("%s:%s", scopename, err)
		return err
	} else {

		sp := Scope{scopename, tfile, outfile, nil}
		sp.Params()
		sp.ProcessTemplate()
	}
	return nil
}

func f(scopename string, ch chan error) {
	err := Execute(scopename)
	ch <- err
}
func main() {
	var err error
	err = config()
	if err != nil {
		log.Fatal("Failed to read config: %s", err)
	}
	scopelist := viper.GetStringSlice("scopes")
	ch := make(chan error, len(scopelist))

	for i, scopename := range scopelist {
		log.Printf("working on scope: %d, %s", i, scopename)

		go f(scopename, ch)
	}
	for i, scopename := range scopelist{
		err = <- ch
		log.Printf("Finished with scope %d:%s", i, scopename)
		if err != nil{
			log.Printf("%s: failed to execute: %s", scopename, err)
		}
	}
}
