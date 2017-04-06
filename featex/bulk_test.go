package featex

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"text/template"
	// "github.com/jpfairbanks/featex"
	_ "github.com/lib/pq"
	"testing"
)

//dbconnect opens the database or fails the test
func dbconnect(t *testing.T) (*sql.DB, error) {
	t.Log("opening DB connection")
	conn, err := sql.Open("postgres", DBString())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("opened DB connection")
	return conn, nil
}

//TestConnection just make sure that we can connect and the tables we expect are there.
func TestConnection(t *testing.T) {
	// conn, err := dbconnect(t)
	conn, err := sql.Open("postgres", DBString())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := conn.Query("select job_id, person_id, concept_id, type from results_lite_synpuf2.features")
	if err != nil {
		t.Fatal(err)
	}
	row := Feature{}
	t.Logf("Feature table")
	var job_id sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&job_id, &row.PersonID, &row.ConceptID, &row.ConceptType)
		if err != nil {
			t.Fatal(err)
		}
		if !job_id.Valid {
			t.Fatal("Job_id of a feature row is null!")
		}
		t.Logf("job_id:%d, row:%v", job_id.Int64, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

// QueryInt runs a query that returns a single int and returns that integer
// useful for running count queries.
func QueryInt(conn *sql.DB, query string, args ...interface{}) (int, error) {
	var res int
	row := conn.QueryRow(query, args...)
	err := row.Scan(&res)
	return res, err
}

// MustQueryInt runs a query that returns a single int and returns that integer fails the test on error
// useful for running a count query as part of a test.
func MustQueryInt(t *testing.T, conn *sql.DB, query string, args ...interface{}) (int, error) {
	res, err := QueryInt(conn, query, args...)
	if err != nil {
		t.Fatal(err)
	}
	return res, err
}

//TestInsertion insert a row of features and then check that it made it.
func TestInsertion(t *testing.T) {
	var job_id int
	var concept_id int = 45943027
	var count int
	conn, err := dbconnect(t)
	q := `insert into results_lite_synpuf2.feature_jobs (created, description) values (now(), 'test job 1') returning job_id`
	job_id, _ = MustQueryInt(t, conn, q)
	q = `insert into results_lite_synpuf2.features (job_id, concept_id, value, type)
                  values ($1, $2, 1, 'drug' )`
	_, err = conn.Query(q, job_id, concept_id)
	if err != nil {
		t.Fatal(err)
	}

	q = `select count(*) from results_lite_synpuf2.features where concept_id=$1`
	count, _ = MustQueryInt(t, conn, q, concept_id)
	t.Logf("Count of matching rows = %d", count)
	if count < 1 {
		t.Fatal("Count does not match")
	}
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

// BulkOptions configuration for the bulk queries.
type BulkOptions struct {
	Schema      string
	Table       string
	JobTable    string
	Positional  int
	Description string
	Selectstmt  string
}

// BulkTemplate: get or create the template for wrapping a select statement with an insert into select clause
// this template requires fields for "positional" and
func BulkTemplate() (*template.Template, error) {
	tpl := template.New("bulk")
	tpl, err := tpl.Parse(`insert into {{.Schema}}.{{.Table}} (job_id, person_id, concept_id, value, type)
                  select ${{.Positional}} as job_id, person_id, concept_id, value, type from ({{.Selectstmt}}) as t;`)
	if err != nil {
		return tpl, err
	}
	return tpl, nil
}

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

// BulkFeatures takes a select query and bulk query options wraps it in the BulkTemplate query
// for inserting into the results table. It first creates a new job in the job table, so that you
// can get the results by filtering the features table by the job_id
func BulkFeatures(conn *sql.DB, query string, opt BulkOptions) (job_id int, err error) {
	var bulk_query string
	var countInsertions string
	bulk_query, err = Wrap(query, opt)
	if err != nil {
		return
	}
	job_id, err = NewJob(conn, opt, opt.Description)
	if err != nil {
		return
	}
	_, err = conn.Query(bulk_query, job_id)
	if err != nil {
		return
	}
	countInsertions, err = CountInsertionsQuery(opt)
	if err != nil {
		return
	}
	count, err := QueryInt(conn, countInsertions, job_id)
	if err != nil {
		return
	}
	var c int = 1
	if count < c {
		err = fmt.Errorf("Insertion did not yield any values")
		return
	}
	return
}

func TestWrap(t *testing.T) {
	var err error
	var kwargs BulkOptions
	q := `select person_id, gender_concept_id as concept_id, 1 as value, 'gender' as type from lite_synpuf2.person
	      UNION
	      select person_id, race_concept_id as concept_id, 1 as value, 'race' as type from lite_synpuf2.person
	      WHERE person.year_of_birth > 1900
	      ORDER BY person_id ASC, type ASC, concept_id ASC, value ASC;`
	kwargs = BulkOptions{
		Schema:      "results_lite_synpuf2",
		Table:       "features",
		JobTable:    "feature_jobs",
		Positional:  1,
		Description: "TestWrap1",
	}
	bulk_query, err := Wrap(q, kwargs)
	t.Logf("wrapped query: %s", bulk_query)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := dbconnect(t)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("connection: %v", conn)
	if conn == nil {
		t.Fatal("Could not connect to DB")
	}
	t.Logf("Creating job")
	job_id, err := NewJob(conn, kwargs, kwargs.Description)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("made job")
	_, err = conn.Query(bulk_query, job_id)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("insertion successful")
	count, err := MustQueryInt(t, conn, "select count(*) from results_lite_synpuf2.features where job_id = $1", job_id)
	if err != nil {
		t.Fatal(err)
	}
	var c int = 10
	if count < c {
		t.Fatalf("Failed to insert at least %d feature rows", c)
	}
	t.Logf("query successful")
}
