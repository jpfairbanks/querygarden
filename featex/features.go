package featex

import (
	"database/sql"
	"time"
	"fmt"
	"github.com/jpfairbanks/featex/log"
	"io"
)

// A Feature is a struct that represents the result of running a feature extraction query.
type Feature struct {
	Personid     sql.NullInt64  //the unique identifier of the row
	Start_date   sql.NullString //time.Time
	End_date     sql.NullString //time.Time
	Concept_id   sql.NullInt64  // the omop concept code that reprsents the feature
	Concept_type string         // a name for the
}

// FmtDate takes a time a sprintf's it as a yyy-mm-dd date.
func FmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

//Println sproa feature as a row of a CSV table.
func (f Feature) Fprintln(w io.Writer) {
	fmt.Fprintf(w, "%d, %s, %s, %d, %s\n", f.Personid.Int64, f.Start_date.String,
		f.End_date.String, f.Concept_id.Int64, f.Concept_type)
}

// CSVRow takes a row of the feature matrix and prints it out as a CSV line to os.Stdout.
// rows is a the result of running a sql.Query
func CSVRow(rows *sql.Rows, w io.Writer) (bool, error) {
	var row Feature
	if err := rows.Scan(&row.Personid, &row.Start_date, &row.End_date, &row.Concept_id, &row.Concept_type); err != nil {
		log.Error(err)
		return false, err
	}
	//spew.Dump("%s", row)
	row.Fprintln(w)
	return true, nil
}