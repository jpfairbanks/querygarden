package featex

import (
	"bytes"
	"database/sql"
	"fmt"
	"testing"

	"github.com/jpfairbanks/featex/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var yamlExample = []byte(`
features:
  demographics: {
    Filename: ../../sql/demographics.sql,
    bindvars: ["limit"] }
  features: {
    Filename: ../../sql/features.sql,
    bindvars: ["person"]}
  demographics_historical:
    Filename: "../../sql/demographics.historical.sql"
    bindvars: []
`)

func TestContext(t *testing.T) {
	// Set up command line flag.
	//Config()
	viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

	var err = viper.ReadConfig(bytes.NewBuffer(yamlExample))
	if err != nil{
		t.Fatal(err)
	}
	s := viper.GetString("features.demographics.Filename")
	if s == "" {
		t.Fatal("could not read configuration file")
	}
	fmt.Printf("test: %s\n", s)
	t.Log(s)
	// Args[0] is the database connection string with the password in it.
	dbstring := DBString()

	// prepare the query map which holds the queries as strings
	ctx := Context{"./sql", make(map[string]Query)}
	log.Info("Loading queries")
	keys := []string{"demographics", "demographics_historical", "features"}
	ctx.LoadQueries(keys)
	//fmt.Printf("%##v", ctx.Queries)
	//spew.Dump(ctx.Queries)

	//prepare the database
	var Conn *sql.DB
	Conn, err = sql.Open("postgres", dbstring)
	if err != nil {
		t.Fatal(err)
	}

	key := "features"
	// execute the query and get a handle to the result
	args := ctx.ArrangeBindVars(key, map[string]string{"person":"2"})
	if args[0] != "2"{
		t.Fatal("Failed to arrange bindvars for query")
	}
	rows, err := ctx.Query(*Conn, key, args...)
	if err != nil {
		log.Fatal(err)
	}
	if rows == nil {
		t.Fatal("rows was nil")
	}

}

func TestQuery_ArrangeBindVars(t *testing.T) {
	q := Query{"./dummy.sql", "Select * from Table where a = ?", []string{"wclause"}}
	args := q.ArrangeBindVars(map[string]string{"wclause":"2"})
	if args[0] !=  "2"{
		t.Fatal("Failed to arrange bindvars for a single query")
	}
}

func TestRowMap(t *testing.T) {

	// Map a function over each row
	//err = RowMap(featex.CSVRow, rows)
	//if err != nil {
	//	log.Fatal(err)
	//}
}