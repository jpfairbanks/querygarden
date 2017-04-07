package featex

import (
	"database/sql"
	"fmt"
	"log"
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

// MustQueryInt runs a query that returns a single int and returns that integer fails the test on error
// useful for running a count query as part of a test.
func MustQueryInt(t *testing.T, conn *sql.DB, query string, args ...interface{}) (int, error) {
	res, err := QueryInt(conn, query, args...)
	if err != nil {
		t.Fatal(err)
	}
	return res, err
}

//TestConnection just make sure that we can connect and the tables we expect are there.
func TestConnection(t *testing.T) {
	// conn, err := dbconnect(t)
	conn, err := sql.Open("postgres", DBString())
	if err != nil {
		t.Fatal(err)
	}
	rows, err := conn.Query("select job_id, person_id, concept_id, type from results_lite_synpuf2.features limit 10")
	if err != nil {
		t.Fatal(err)
	}
	row := Feature{}
	t.Logf("Feature table")
	var job_id sql.NullInt64
	count := 0
	for rows.Next() {
		err := rows.Scan(&job_id, &row.PersonID, &row.ConceptID, &row.ConceptType)
		if err != nil {
			t.Fatal(err)
		}
		if !job_id.Valid {
			t.Fatal("Job_id of a feature row is null!")
		}
		t.Logf("job_id:%d, row:%v", job_id.Int64, row)
		count += 1
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if count != 10 {
		t.Fatalf("Incorrect number of rows returned (limit 10): %d", count)
	}
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

func ExecuteBulk(con *sql.DB, query string, opt BulkOptions) (job_id int, err error) {
	job_id, err = BulkFeatures(con, query, opt)
	if err != nil {
		return
	}
	if job_id < 1 {
		return 0, fmt.Errorf("Bad Job_id < 1")
	}
	return
}

func TestWrap(t *testing.T) {
	var err error
	kwargs := BulkOptions{
		Schema:      "results_lite_synpuf2",
		Table:       "features",
		JobTable:    "feature_jobs",
		Positional:  1,
		Description: "",
	}

	conn, err := dbconnect(t)
	if err != nil {
		t.Fatal(err)
	}
	queries := []string{OMOPQ.demographic, OMOPQ.drug, OMOPQ.procedure, OMOPQ.condition}
	descriptions := []string{"demoquery", "drugquery", "procedurequery", "conditionquery"}
	for i, q := range queries {
		kwargs.Description = descriptions[i]
		job_id, err := ExecuteBulk(conn, q, kwargs)
		if err != nil {
			t.Fatal(err)
		}
		log.Printf("job_id: %d", job_id)
	}
}
