package depr

import (
	"io"
	"log"
	"text/template"
	"errors"
	"io/ioutil"
	"os"
)


func OpenOutput(args []string) (io.Writer, error) {
	var outfile io.Writer
	var err error
	if len(args) > 1 {
		log.Printf("Writing Rendered Template to file:%s", args[1])
		outfile, err = os.Create(args[1])
	}
	if err != nil {
		log.Println("Cannot open output file supplied: %v", err)
		return outfile, err
	}
	return outfile, nil
}

func ReadStringName(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename) // just pass the file name
	if err != nil {
		log.Print(err)
		return "", err
	}
	str := string(b)
	return str, nil
}
func ProcessTemplate(output io.Writer, filename string, params interface{}) error {
	t := template.New(filename)
	s, err := ReadStringName(filename)
	if err != nil {
		log.Printf("Could not read file %s\n", filename)
		return err
	}
	t, err = t.Parse(s)
	if err != nil {
		log.Println("Failed to parse: %s\n", err)
		return err
	}
	if output == nil{
		return  errors.New("Output io.Writer was nil")
	}
	err = t.Execute(output, params)
	if err != nil {
		log.Printf("Failed to execute: %s\n", err)
	}
	return err
}

func main() {
}
