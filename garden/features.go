package garden

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/jpfairbanks/featex/log"
)

// A Feature is a struct that represents the result of running a feature extraction query.
type Feature struct {
	PersonID    sql.NullInt64  // the unique identifier of the row
	StartDate   sql.NullString // time.Time
	EndDate     sql.NullString // time.Time
	ConceptID   sql.NullInt64  // the omop concept code that reprsents the feature
	ConceptType string         // a name for the
}

// FmtDate takes a time a sprintf's it as a yyy-mm-dd date.
func FmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// Fprintln writes a feature as a row of a CSV table.
// Use with a bufio.Writer to build up the tables
func (f Feature) Fprintln(w io.Writer) {
	fmt.Fprintf(w, "%d, %s, %s, %d, %s\n", f.PersonID.Int64, f.StartDate.String,
		f.EndDate.String, f.ConceptID.Int64, f.ConceptType)
}

// CSVRow takes a row of the feature matrix and prints it out as a CSV line to os.Stdout.
// rows is a the result of running a sql.Query
func CSVRow(rows *sql.Rows, w io.Writer) (bool, error) {
	var row Feature
	if err := rows.Scan(&row.PersonID, &row.StartDate, &row.EndDate, &row.ConceptID, &row.ConceptType); err != nil {
		log.Error(err)
		return false, err
	}
	//spew.Dump("%s", row)
	row.Fprintln(w)
	return true, nil
}
