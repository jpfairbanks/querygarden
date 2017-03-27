package main

import (
	"database/sql"
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"encoding/json"
	"github.com/jpfairbanks/featex/featex"
	"github.com/jpfairbanks/featex/log"
	_ "github.com/lib/pq"
	"net/http"
	"io"

	//"html/template"
)

var RESPONSE_LIMIT = 250
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
		i += 1
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

func WriteString(w io.Writer, s string) (int, error) {
	return w.Write([]byte(s))
}
var headelt = `<head>
<link href="https://maxcdn.bootstrapcdn.com/bootswatch/3.3.7/flatly/bootstrap.min.css" rel="stylesheet" integrity="sha384-+ENW/yibaokMnme+vBLnHMphUYxHs34h9lpdbSLuAwGkOKFRl4C34WkjazBtb7eT" crossorigin="anonymous">
<link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.10.0/styles/default.min.css">
<script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.10.0/highlight.min.js"></script>
<script>hljs.initHighlightingOnLoad();</script>
</head>`

var tableheader = `<h2>Result</h2><p><table class="table table-stripped"><tr>Person, Start Date, End Date, ConceptID, Feature Type</tr>`
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
	ctx := featex.Context{"./sql", make(map[string]featex.Query)}
	log.Info("Loading queries")
	keys := []string{"demographics", "demographics_historical", "features"}
	ctx.LoadQueries(keys)

	//prepare the database
	var Conn *sql.DB
	Conn, err = sql.Open("postgres", featex.DBString())
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Database connected")

	// set up the query handler with the current Context and DB connection
	queryhandler := func(w http.ResponseWriter, r *http.Request) {
		req, err := featex.ParseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		args := make(map[string]string)
		for k, v := range req.Args {
			args[k] = v[0]
		}
		byts, err := json.Marshal(req)
		WriteString(w, headelt)
		WriteString(w,`<div class="container"><h1>Ran Query</h1>`)
		WriteString(w, "<p><pre><code class=json>")
		w.Write(byts)
		WriteString(w, "</code></pre></p>")

		WriteString(w, "<h2>Query.Text</h2><pre><code class=sql>")
		WriteString(w, ctx.Queries[req.Key].Text)
		WriteString(w, "</code></pre>")
		WriteString(w, tableheader)

		rows, err := ctx.Query(*Conn, req.Key, ctx.ArrangeBindVars(req.Key, args)...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var i = 0
		RowMap(func(r *sql.Rows) (bool, error) {
			i += 1
			if i > RESPONSE_LIMIT {
				return false, nil
			} else {
				WriteString(w, "<tr><td>")
				b, err := featex.CSVRow(rows, w)
				WriteString(w, "</td></tr>")
				return b,err
			}

			return true, nil

		},
			rows)
		WriteString(w, "</table></p></div>")
	}

	//Attach the query handler to the route and start the server on localhost.
	http.HandleFunc("/query/", queryhandler)
	addr := ":8080"
	log.Infof("Serving on address: %s", addr)
	http.ListenAndServe(addr, nil)
}
