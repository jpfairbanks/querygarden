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
	personid     sql.NullInt64  //the unique identifier of the row
	start_date   sql.NullString //time.Time
	end_date     sql.NullString //time.Time
	concept_id   sql.NullInt64  // the omop concept code that reprsents the feature
	concept_type string         // a name for the
}

// FmtDate takes a time a sprintf's it as a yyy-mm-dd date.
func FmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

//Println sproa feature as a row of a CSV table.
func (f Feature) Fprintln(w io.Writer) {
	fmt.Fprintf(w, "%d, %s, %s, %d, %s\n", f.personid.Int64, f.start_date.String,
		f.end_date.String, f.concept_id.Int64, f.concept_type)
}

// CSVRow takes a row of the feature matrix and prints it out as a CSV line to os.Stdout.
// rows is a the result of running a sql.Query
func CSVRow(rows *sql.Rows, w io.Writer) (bool, error) {
	var row Feature
	if err := rows.Scan(&row.personid, &row.start_date, &row.end_date, &row.concept_id, &row.concept_type); err != nil {
		log.Error(err)
		return false, err
	}
	//spew.Dump("%s", row)
	row.Fprintln(w)
	return true, nil
}