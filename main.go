package main

import (
	"database/sql"
	"flag"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/jpfairbanks/featex/featex"
	"github.com/jpfairbanks/featex/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"time"
)

//go:generate sqlgen

func config() error {
	viper.SetConfigName("featex_config")
	viper.AddConfigPath("/etc/featex/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.featex") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	defglobal := map[string]interface{}{
		"schema":  "schema",
		"version": "1.0",
	}
	viper.SetDefault("global", defglobal)
	err := viper.ReadInConfig()
	if err != nil {
		log.Error(err)
		panic("Cannot load configuration check for file featex.config in ./, /etc/featex/, or $HOME/.featex/")
	}
	return nil
}

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
func (f Feature) Println() {
	fmt.Printf("%d, %s, %s, %d, %s\n", f.personid.Int64, f.start_date.String,
		f.end_date.String, f.concept_id.Int64, f.concept_type)
}

// CSVRow takes a row of the feature matrix and prints it out as a CSV line to os.Stdout.
// rows is a the result of running a sql.Query
func CSVRow(rows *sql.Rows) error {
	var row Feature
	if err := rows.Scan(&row.personid, &row.start_date, &row.end_date, &row.concept_id, &row.concept_type); err != nil {
		log.Error(err)
		return err
	}
	//spew.Dump("%s", row)
	row.Println()
	return nil
}

// RowMap takes a DB result Rows and maps a function over each row.
// This function handles the errors by logging and breaking out of the loop.
// The rows object is left open so that you can get the data out of it
// if you want to do that.
func RowMap(f func(rows *sql.Rows) error, rows *sql.Rows) error {
	var i int
	i = 0
	for rows.Next() {
		err := f(rows)
		if err != nil {
			log.Error("scan error row number: ", i)
			return err
		}
		i += 1
	}
	if err := rows.Err(); err != nil {
		log.Error("error post gather row number: ", i)
		log.Fatal(err)
		return err
	}
	log.Infof("gathered %d rows", i)
	return nil

}

func main() {
	// Set up command line flag.
	config()
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
	// Args[0] is the database connection string with the password in it.
	dbstring := flag.Arg(0)

	// prepare the query map which holds the queries as strings
	ctx := featex.Context{"./sql", make(map[string]featex.Query)}
	log.Info("Loading queries")
	keys := []string{"demographics", "demographics_historical", "features"}
	ctx.LoadQueries(keys)
	//fmt.Printf("%##v", ctx.Queries)
	//spew.Dump(ctx.Queries)

	//prepare the database
	var Conn *sql.DB
	if dbstring == "" {
		dbstring = "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full"
	}
	Conn, err := sql.Open("postgres", dbstring)
	if err != nil {
		log.Fatal(err)
	}

	key := "features"
	// execute the query and get a handle to the result
	args := ctx.ArrangeBindVars(key, map[string]string{"person":"2"})

	rows, err := ctx.Query(*Conn, key, args...)
	if err != nil {
		log.Fatal(err)
	}

	// Map a function over each row
	err = RowMap(CSVRow, rows)
	if err != nil {
		log.Fatal(err)
	}
}
