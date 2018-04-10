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
	"os"

	"github.com/jpfairbanks/querygarden/garden"
	"github.com/jpfairbanks/querygarden/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"strings"
)

// ResponseLimit is the maximum number of values to pass as an HTML table
var ResponseLimit = 500

// Global variable of templates so that all functions can share the same templates
var templates *template.Template

//go:generate querygen

// RowMap takes a DB result Rows and maps a function over each row.
// This function handles the errors by logging and breaking out of the loop.
// The rows object is left open so that you can get the data out of it
// if you want to do that.
func RowMap(f func(rows *sql.Rows) (bool, error), rows *sql.Rows) error {
	if rows == nil {
		return fmt.Errorf("Rows argument is nil")
	}
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
func WriteTable(tablew io.Writer, rows *sql.Rows) (int, []string, error) {
	var i = 0
	// var value int
	columns, err := rows.Columns()
	if err != nil {
		return i, columns, err
	}
	dest := make([]interface{}, len(columns))
	brow := make([][]byte, len(columns))
	for i, _ := range brow {
		dest[i] = &brow[i]
	}
	err = RowMap(func(r *sql.Rows) (bool, error) {
		i++
		if i > ResponseLimit {
			// false means to terminate
			return false, nil
		}
		// get the value out of the db cursor
		if err = rows.Scan(dest...); err != nil {
			log.Error(err)
			return false, err
		}
		// write one row to the output stream
		fmt.Fprintf(tablew, "<tr>")
		for _, r := range brow {
			fmt.Fprintf(tablew, "<td>%s</td>", bytes.Trim(r, "\n"))
		}
		fmt.Fprintf(tablew, "</tr>")
		// true means continue
		return true, nil

	}, rows)
	if err != nil {
		return i, columns, err
	}
	return i, columns, nil
}

//TakeFirst takes only the first argument for each key in the map.
func TakeFirst(req garden.Request) map[string]string {
	args := make(map[string]string)
	for k, v := range req.Args {
		args[k] = v[0]
	}
	return args
}

func htmlerrorpage(w http.ResponseWriter, err error, code int) {
	log.Debug("Writing an error message as html")
	log.Debug(err.Error())
	w.WriteHeader(code)
	d := map[string]interface{}{"StatusText": http.StatusText(code), "Msg": err.Error()}
	err = templates.ExecuteTemplate(w, "500.html.tmpl", d)
	if err != nil {
		log.Error(err)
	}
}

func fivehundred(w http.ResponseWriter, err error) {
	htmlerrorpage(w, err, http.StatusInternalServerError)
}

func LoadTemplates() {
	// load the templates into a template cache panic on error.
	templates = template.Must(template.ParseFiles("templates/html/queries.html.tmpl",
		"templates/html/query.html.tmpl",
		"templates/html/index.html.tmpl",
		"templates/html/404.html.tmpl",
		"templates/html/500.html.tmpl",
		"templates/html/login.html.tmpl",
		"templates/html/reload.html.tmpl",
	))
	log.WithFields(log.Fields{"Templates": templates.DefinedTemplates()}).Info("Read Templates")
}

// LoadQueries: prepare the global query map which holds the queries as strings
func LoadQueries(ctx garden.Context, configkey string) (keys []string, err error) {
	log.Info("Loading queries")

	featuresconf := viper.GetStringMap("features")
	keys = make([]string, len(featuresconf))
	i := 0
	for k := range featuresconf {
		keys[i] = k
		i++
	}
	log.Infof("The keys are: %v", keys)
	err = ctx.LoadQueries(keys)
	return
}

func ForceLogin(w http.ResponseWriter, r *http.Request) {
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", 302)
	}
}

func main() {
	// Set up command line flag.
	err := garden.Config("featex")
	resultsSchema := viper.GetString("global.rschema")
	log.Infof("resultsSchema=%s", resultsSchema)
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
	ctx := garden.Context{Querypath: "./sql", Queries: make(map[string]garden.Query)}
	_, err = LoadQueries(ctx, "features")

	if err != nil {
		log.Fatal("Cannot load queries: ABORT!")
	}
	//prepare the database
	var Conn *sql.DB
	Conn, err = sql.Open("postgres", garden.DBString())
	log.Debug(garden.DBString())
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Database connected")

	LoadTemplates()

	// set up the query handler with the current Context and DB connection
	queryhandler := func(w http.ResponseWriter, r *http.Request) {
		ForceLogin(w, r)
		// initialize a buffer for the table in HTML
		var nrows int
		buf := make([]byte, 0)
		table := bytes.NewBuffer(buf)
		tablew := bufio.NewWriter(table)

		// parse the request from the client
		req, err := garden.ParseRequest(r)
		if err != nil {
			htmlerrorpage(w, err, http.StatusBadRequest)
			return
		}
		args := TakeFirst(req)

		// get the result from the DB
		params, err := ctx.ArrangeBindVars(req.Key, args)
		if err != nil {
			htmlerrorpage(w, err, http.StatusBadRequest)
			return
		}
		rows, err := ctx.Query(Conn, req.Key, params...)
		if err != nil {
			fivehundred(w, err)
			return
		}
		defer rows.Close()
		// write the result as a table to the a bytes buffer so it can go in the template.
		var headrow []string
		nrows, headrow, err = WriteTable(tablew, rows)
		log.Debugf("processed %d rows", nrows)
		if err != nil {
			fivehundred(w, err)
			return
		}
		err = tablew.Flush()
		if err != nil {
			//fivehundred(w, err)
			return
		}
		byts, err := json.Marshal(req)
		log.Debugf("%s", byts)
		var respdata = map[string]interface{}{"Args": req,
			"QueryText": ctx.Queries[req.Key].Text,
			"Table":     template.HTML(table.Bytes()),
			"Headrow":   headrow,
			"Bindvars":  ctx.Queries[req.Key].Bindvars}
		// render the page to the client
		err = templates.ExecuteTemplate(w, "query.html.tmpl", respdata)
		if err != nil {
			err = fmt.Errorf("Could not render template: %s", err)
			fivehundred(w, err)
		}
	}
	//Attach the query handler to the route and start the server on localhost.
	//http.Handle("/", http.RedirectHandler("/index.html", http.StatusPermanentRedirect))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{"": ""}
		b := &bytes.Buffer{}
		err := templates.ExecuteTemplate(b, "404.html.tmpl", data)
		w.Write(b.Bytes())
		if err != nil {
			fivehundred(w, err)
		}
		//http.Error(w, b.String(), http.StatusNotFound)
	})
	http.HandleFunc("/query/", queryhandler)
	listhandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(ctx.Queries)
		//resp, err := json.Marshal(ctx.Queries)
		if err != nil {
			fivehundred(w, fmt.Errorf("Could not marshall ctx.Queries into JSON"))
		}

	}
	listhandlerhtml := func(w http.ResponseWriter, r *http.Request) {
		err := templates.ExecuteTemplate(w, "queries.html.tmpl", ctx)
		if err != nil {
			fivehundred(w, fmt.Errorf("Could not render queries as html %s", err.Error()))
		}

	}
	http.HandleFunc("/queries.json", listhandler)
	http.HandleFunc("/queries.html", listhandlerhtml)
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{"": ""}
		err := templates.ExecuteTemplate(w, "index.html.tmpl", data)
		if err != nil {
			fivehundred(w, err)
		}
	})

	bulkhandler := func(w http.ResponseWriter, r *http.Request) {
		//var err error
		var resp *garden.BulkResponse = new(garden.BulkResponse)
		w.Header().Set("Content-Type", "application/json")
		route := r.URL.Path
		parts := strings.Split(route, "/")
		key := parts[2]
		qry, ok := ctx.Queries[key]
		if !ok {
			htmlerrorpage(w, fmt.Errorf("Could not find key=%s", key), http.StatusBadRequest)
			log.Debugf("Available Keys are: %v")
			for k := range ctx.Queries {
				log.Debugf("\t%s", k)
			}
			return
		}

		resp.SourceQuery = qry
		log.WithFields(log.Fields{"key": key,
			"route": route,
			"query": qry}).Debug("Loaded Query for Bulk")

		// getting the query arguments
		req, err := garden.ParseRequest(r)
		if err != nil {
			htmlerrorpage(w, err, http.StatusBadRequest)
			log.WithFields(log.Fields{"key": key, "status": http.StatusBadRequest,
				"error": err}).Error("failed to parse request")
			return
		}
		qargs := TakeFirst(req)

		log.WithFields(log.Fields{"key": key, "qargs": qargs}).Debug("parsed query arguments")

		opts := garden.BulkOptions{Schema: resultsSchema,
			Table:       "features",
			JobTable:    "feature_jobs",
			Positional:  len(qry.Bindvars) + 1,
			Description: "api query key=" + key,
			Selectstmt:  qry.Text}
		resp.BulkOptions = opts
		// q, err := garden.RenderTemplate(garden.BulkTemps.Insert, opts)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }
		// log.WithFields(log.Fields{"insertquery": q}).Info("Insertion into bulk table")

		log.Debug("calling BulkFeatures")
		var jobID int
		//args := ctx.ArrangeBindVars(key, qargs)
		args, err := ctx.ArrangeBindVars(key, qargs)
		if err != nil {
			htmlerrorpage(w, err, http.StatusBadRequest)
			return
		}
		log.WithFields(log.Fields{"key": key, "args": args}).Info("Bulk query arguments:")
		jobID, err = garden.BulkFeatures(Conn, qry.Text, opts, args...)
		resp.JobID = jobID
		resp.Err = err
		if err != nil {
			log.WithFields(log.Fields{"key": key, "status": 500, "err": err}).Error("Could not make bulk query")
			fivehundred(w, fmt.Errorf("Query Failed: %v", err))
		}
		resp.SelectStmt = fmt.Sprintf("select * from %s.%s where job_id=%d",
			opts.Schema, opts.Table, resp.JobID)
		js, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			log.WithFields(log.Fields{"key": key, "status": 500}).Error("Could not marshall json response")
			fivehundred(w, err)
		}
		n, err := w.Write(js)
		if err != nil {
			log.WithFields(log.Fields{"key": key, "error": err, "bytes": n}).Error("Failed to write response")
			return
		}

	}
	http.HandleFunc("/bulk/", bulkhandler)
	http.HandleFunc("/login", loginhandler)
	http.HandleFunc("/reload", func(w http.ResponseWriter, r *http.Request) {
		ForceLogin(w, r)
		vars := make(map[string]interface{})
		LoadTemplates()
		names := make([]string, len(templates.Templates()))
		for i, t := range templates.Templates() {
			names[i] = t.Name()
		}
		vars["Templates"] = names
		keys, err := LoadQueries(ctx, "features")
		if err != nil {
			fivehundred(w, err)
		}
		vars["Queries"] = keys
		err = templates.ExecuteTemplate(w, "reload.html.tmpl", vars)
		if err != nil {
			fivehundred(w, err)
		}
	})
	addr, ok := os.LookupEnv("FEATEX_ADDR")
	if !ok {
		addr = "0.0.0.0:8080"
	}
	log.Infof("Serving on address: %s", addr)
	if os.Getenv("FEATEX_TLS") != "" {
		err = http.ListenAndServeTLS(addr, "server.crt", "server.key", nil)
	} else {
		err = http.ListenAndServe(addr, nil)
	}
	if err != nil {
		log.Fatalf("Error in serving process is dying:\n\t%s\n%q", err.Error(), err)
	}
}
