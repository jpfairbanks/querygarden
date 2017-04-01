package main

import (
	"database/sql"
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"bufio"
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"

	"github.com/jpfairbanks/featex/featex"
	"github.com/jpfairbanks/featex/log"
	_ "github.com/lib/pq"
)

// ResponseLimit is the maximum number of values to pass as an HTML table
var ResponseLimit = 500

//go:generate sqlgen

// RowMap takes a DB result Rows and maps a function over each row.
// This function handles the errors by logging and breaking out of the loop.
// The rows object is left open so that you can get the data out of it
// if you want to do that.
func RowMap(f func(rows *sql.Rows) (bool, error), rows *sql.Rows) error {
	var i int
	i = 0
	for rows.Next() {
		cont, err := f(rows)
		if err != nil {
			log.Error("scan error row number: ", i)
			return err
		}
		i++
		if !cont {
			break
		}
	}
	if err := rows.Err(); err != nil {
		log.Error("error post gather row number: ", i)
		log.Fatal(err)
		return err
	}
	log.Infof("gathered %d rows", i)
	return nil

}

// WriteString writes out a string as if it were a []byte
func WriteString(w io.Writer, s string) (int, error) {
	return w.Write([]byte(s))
}

// WriteTable writes a series of rows as a table of featex.Features.
// The io.Writer should be a buffer of some kind that will get passed to
// an html template as a template.HTML(tablew.Bytes()) object.
// Otherwise, the table tags will be escaped and break your page.
// Make sure to flush the buffer before constructing the template.HTML object.
// returns the number of rows successfully writen
func WriteTable(tablew io.Writer, rows *sql.Rows) (int, error) {
	var i = 0
	var row featex.Feature
	err := RowMap(func(r *sql.Rows) (bool, error) {
		i++
		if i > ResponseLimit {
			// false means to terminate
			return false, nil
		}
		// get the value out of the db cursor
		if err := rows.Scan(&row.PersonID, &row.StartDate, &row.EndDate, &row.ConceptID, &row.ConceptType); err != nil {
			log.Error(err)
			return false, err
		}
		// write one row to the output stream
		fmt.Fprintf(tablew, "<tr><td>%d</td> <td>%s</td> <td> %s</td> <td> %d</td> <td> %s</td></tr>\n",
			row.PersonID.Int64, row.StartDate.String,
			row.EndDate.String, row.ConceptID.Int64, row.ConceptType)
		// true means continue
		return true, nil

	}, rows)
	if err != nil {
		return i, err
	}
	return i, nil
}

var headelt = template.HTML(`<head>
<link href="https://maxcdn.bootstrapcdn.com/bootswatch/3.3.7/flatly/bootstrap.min.css" rel="stylesheet" integrity="sha384-+ENW/yibaokMnme+vBLnHMphUYxHs34h9lpdbSLuAwGkOKFRl4C34WkjazBtb7eT" crossorigin="anonymous">
<link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.10.0/styles/default.min.css">
<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.10.0/highlight.min.js"></script>
<script>hljs.initHighlightingOnLoad();</script>
</head>`)

var tableheader = template.HTML(`<table class="table table-stripped"><tr>Person, Start Date, End Date, ConceptID, Feature Type</tr>`)

func main() {
	// Set up command line flag.
	err := featex.Config()
	if err != nil {
		panic(err)
	}
	versionFlag := flag.Bool("version", false, "Version")
	flag.Parse()
	if *versionFlag {
		fmt.Println("Git Commit:", GitCommit)
		fmt.Println("Version:", Version)
		if VersionPrerelease != "" {
			fmt.Println("Version PreRelease:", VersionPrerelease)
		}
		return
	}

	// prepare the query map which holds the queries as strings
	ctx := featex.Context{Querypath: "./sql", Queries: make(map[string]featex.Query)}
	log.Info("Loading queries")
	keys := []string{"demographics", "demographics_historical", "features", "condition", "drugs", "drug_era", "milenial_features"}
	ctx.LoadQueries(keys)

	//prepare the database
	var Conn *sql.DB
	Conn, err = sql.Open("postgres", featex.DBString())
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Database connected")

	// load the templates into a template cache panic on error.
	var templates = template.Must(template.ParseFiles("templates/html/queries.html.tmpl", "templates/html/query.html.tmpl", "templates/html/index.html.tmpl"))
	log.WithFields(log.Fields{"Templates": templates}).Info("Read Templates")

	// set up the query handler with the current Context and DB connection
	queryhandler := func(w http.ResponseWriter, r *http.Request) {
		// initialize a buffer for the table in HTML
		buf := make([]byte, 0)
		table := bytes.NewBuffer(buf)
		tablew := bufio.NewWriter(table)

		// parse the request from the client
		req, err := featex.ParseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		args := make(map[string]string)
		for k, v := range req.Args {
			args[k] = v[0]
		}

		// get the result from the DB
		rows, err := ctx.Query(Conn, req.Key, ctx.ArrangeBindVars(req.Key, args)...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// write the result as a table to the a bytes buffer so it can go in the template.
		WriteTable(tablew, rows)
		tablew.Flush()
		byts, err := json.Marshal(req)
		var respdata = map[string]interface{}{"Args": string(byts),
			"QueryText":   ctx.Queries[req.Key].Text,
			"Tableheader": tableheader,
			"Table":       template.HTML(table.Bytes())}
		// render the page to the client
		err = templates.ExecuteTemplate(w, "query.html.tmpl", respdata)
		if err != nil {
			err = fmt.Errorf("Could not render template: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	//Attach the query handler to the route and start the server on localhost.
	http.HandleFunc("/query/", queryhandler)
	listhandler := func(w http.ResponseWriter, r *http.Request) {
		resp, err := json.Marshal(ctx.Queries)
		if err != nil{
			http.Error(w, "Could not marshall ctx.Queries into JSON", http.StatusInternalServerError)
		}
		w.Write(resp)
	}
	http.HandleFunc("/queries", listhandler)
	addr := ":8080"
	log.Infof("Serving on address: %s", addr)
	http.ListenAndServe(addr, nil)
}
