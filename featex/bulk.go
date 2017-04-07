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
	Schema      string
	Table       string
	JobTable    string
	Positional  int
	Description string
	Selectstmt  string
}

// NewJob execute the query for making a new feature jobs.
func NewJob(conn *sql.DB, opts BulkOptions, args ...interface{}) (int, error) {
	var s string
	tpl := template.New("job")
	tpl, err := tpl.Parse(`insert into {{.Schema}}.{{.JobTable}} (created, description) values (now(), $1) returning job_id`)
	var q bytes.Buffer
	tpl.Execute(&q, opts)
	s = q.String()
	log.Printf("job query: %s", s)
	job_id, err := QueryInt(conn, s, args...)
	return job_id, err
}

// BulkTemplate: get or create the template for wrapping a select statement with an insert into select clause
// this template requires fields for "positional" and
func BulkTemplate() (*template.Template, error) {
	tpl := template.New("bulk")
	tpl, err := tpl.Parse(`insert into {{.Schema}}.{{.Table}} (job_id, person_id, concept_id, value, type, start_date, end_date)
                  select ${{.Positional}} as job_id, person_id, concept_id, value, type, start_date, end_date from ({{.Selectstmt}}) as t;`)
	if err != nil {
		return tpl, err
	}
	return tpl, nil
}

// CountInsertionsQuery get the query string for counting the number of features associated with a job_id.
func CountInsertionsQuery(opt BulkOptions) (query string, err error) {
	var q bytes.Buffer
	tpl := template.New("countInsertions")
	tpl, err = tpl.Parse("select count(*) from {{.Schema}}.{{.Table}} where job_id = $1")
	if err != nil {
		return
	}
	err = tpl.Execute(&q, opt)
	if err != nil {
		return
	}
	return q.String(), nil
}

// Wrap: converts a select statement into a query that inserts the results into the results table.
// This allows users to define results in terms of the select query that they would write in order
// to retrieve the data and the system will convert this into an insertion to the results table.
func Wrap(query string, kwargs BulkOptions) (string, error) {
	var s string
	var err error
	var q bytes.Buffer
	var tpl *template.Template

	kwargs.Selectstmt = strings.TrimRight(query, ";")

	tpl, err = BulkTemplate()
	log.Printf("template: %v", tpl)
	err = tpl.Execute(&q, kwargs)
	if err != nil {
		return s, err
	}
	s = q.String()
	return s, nil
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
	var bulk_query string      // the query that does bulk insertions
	var countInsertions string // a query to count that we did insertions
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
	job_id, err = NewJob(db, opt, opt.Description)
	if err != nil {
		return
	}
	conn := ErrConnection{Conn: db}
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
