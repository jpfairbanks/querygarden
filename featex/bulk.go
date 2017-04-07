package featex

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"text/template"
)

// BulkOptions configuration for the bulk queries.
type BulkOptions struct {
	Schema      string // Schema for storing results
	Table       string // Table for storing results
	JobTable    string // Table for storing the job_ids and descriptions
	Positional  int    // The number of positional arguments in the Selectstmt
	Description string // A human readable description of the bulk job
	Selectstmt  string // The select statement that gets the data you want to store in the results table
}

// BulkTemplates is a struct for storing all the templates required for conducting a bulk query.
// These templates allow the BulkOptions to configure the operation of the bulk insertions at run time.
type BulkTemplates struct {
	NewJob          *template.Template // Inserts a job into the job table returning the foreign key
	Insert          *template.Template // Inserts the data into the results table
	CountInsertions *template.Template // Counts the amount of inserted data
}

func NewBulkTemplates() (bt BulkTemplates) {
	var tpl *template.Template
	var err error
	tpl = template.New("job")
	tpl, err = tpl.Parse(`insert into {{.Schema}}.{{.JobTable}} (created, description) values (now(), $1) returning job_id`)
	if err != nil {
		panic("Could not parse templates" + err.Error())
	}
	bt.NewJob = tpl
	tpl = template.New("bulk")
	tpl, err = tpl.Parse(`insert into {{.Schema}}.{{.Table}} (job_id, person_id, concept_id, value, type, start_date, end_date)
                  select ${{.Positional}} as job_id, person_id, concept_id, value, type, start_date, end_date from ({{.Selectstmt}}) as t;`)
	if err != nil {
		panic("Could not parse templates" + err.Error())
	}
	bt.Insert = tpl
	tpl = template.New("countInsertions")
	tpl, err = tpl.Parse("select count(*) from {{.Schema}}.{{.Table}} where job_id = $1")
	if err != nil {
		panic("Could not parse templates" + err.Error())
	}
	bt.CountInsertions = tpl
	return
}

var bulkTemplates BulkTemplates = NewBulkTemplates()

// RenderTemplate execute a template with BulkOptions
func RenderTemplate(tpl *template.Template, opts BulkOptions) (string, error) {
	var s string
	var q bytes.Buffer
	err := tpl.Execute(&q, opts)
	s = q.String()
	return s, err
}

// NewJob get the query for making a new feature jobs.
func NewJob(opts BulkOptions) (string, error) {
	tpl := bulkTemplates.NewJob
	s, err := RenderTemplate(tpl, opts)
	log.Printf("job query: %s", s)
	return s, err
}

// CountInsertionsQuery get the query string for counting the number of features associated with a job_id.
func CountInsertionsQuery(opt BulkOptions) (query string, err error) {
	tpl := bulkTemplates.CountInsertions
	return RenderTemplate(tpl, opt)
}

// Wrap: converts a select statement into a query that inserts the results into the results table.
// This allows users to define results in terms of the select query that they would write in order
// to retrieve the data and the system will convert this into an insertion to the results table.
func Wrap(query string, kwargs BulkOptions) (string, error) {
	var tpl *template.Template

	kwargs.Selectstmt = strings.TrimRight(query, ";")

	tpl = bulkTemplates.Insert
	return RenderTemplate(tpl, kwargs)
}

// ErrConnection a connection with an error field so that you can make multiple
// queries on the same connection without checking the error every time.
// This is based on the "Errors are Values" blog post errWriter struct.
type ErrConnection struct {
	Conn *sql.DB
	Err  error
}

// Query just like database/sql DB.Query but with built in error handling.
func (ec *ErrConnection) Query(query string, args ...interface{}) *sql.Rows {
	if ec.Err != nil {
		return nil
	}
	rows, err := ec.Conn.Query(query, args...)
	if err != nil {
		ec.Err = err
	}
	return rows
}

// Query just like database/sql DB.Query but with built in error handling.
func (ec *ErrConnection) QueryInt(query string, args ...interface{}) int {
	var i int
	if ec.Err != nil {
		return i
	}
	i, err := QueryInt(ec.Conn, query, args...)
	if err != nil {
		ec.Err = err
	}
	return i
}

// BulkFeatures takes a select query and bulk query options wraps it in the BulkTemplate query
// for inserting into the results table. It first creates a new job in the job table, so that you
// can get the results by filtering the features table by the job_id
func BulkFeatures(db *sql.DB, query string, opt BulkOptions) (job_id int, err error) {
	// make queries
	var job_query string       // query to add a job
	var bulk_query string      // the query that does bulk insertions
	var countInsertions string // a query to count that we did insertions
	job_query, err = NewJob(opt)
	if err != nil {
		return
	}
	bulk_query, err = Wrap(query, opt)
	if err != nil {
		return
	}
	countInsertions, err = CountInsertionsQuery(opt)
	if err != nil {
		return
	}

	// Modify the database.
	// Insert a new job into the job table
	conn := ErrConnection{Conn: db}
	job_id = conn.QueryInt(job_query, opt.Description)
	log.Printf("bulk_query: %s", bulk_query)
	// do the insertions
	_ = conn.Query(bulk_query, job_id)
	// count how many we did
	count := conn.QueryInt(countInsertions, job_id)
	err = conn.Err
	if err != nil {
		return
	}
	// check that there were insertions.
	var c int = 1
	if count < c {
		err = fmt.Errorf("Insertion did not yield any values")
		return
	}
	return
}
