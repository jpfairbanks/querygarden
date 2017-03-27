package featex

import (
	"github.com/jpfairbanks/featex/log"
	"os"
	"path"
	"io/ioutil"
	"database/sql"
	"fmt"
	"github.com/spf13/viper"
)

// QError tracks which key we were operating on when the Error occured
type QError struct {
	key string
	msg string
}

func (err QError) Error() string  {
	return fmt.Sprintf("key=%s: %msg", err.key, err.msg)
}

// A Query is a structure that holds the
type Query struct {
	Filename string    // the filename from which the query was read
	Text	string     // the raw text of the query
	Bindvars []string  // the names of the query parameters in order expected by the DB

}

// ArrangeBindVars takes a map containing values you want to bind into the parameters of the query
// and arranges them in the order necessary for the sqldriver to process them. Uses default values of the go type
// if no value is found for any key.
func (q Query) ArrangeBindVars(values map[string]string) []interface{}{
	args := make([]interface{}, len(q.Bindvars))
	for i, bv := range q.Bindvars{
		val, ok := values[bv]
		if !ok {
			log.Warnf("Missing Var: pos=%d varname=%s ", i, bv)
		}
		args[i] = val
	}
	return args
}

// A Context holds all the queries in a map keyed by their names
type Context struct {
	Querypath string         // the default directory for searching for queries
	Queries map[string]Query // names -> Query structs
}

// LoadQueries from the filesystem given a list of keys.
func (ctx *Context) LoadQueries(keys []string) error {
	log.Info(ctx.Querypath)
	for _, key := range keys{
		q, err := ctx.LoadQuery(key)
		if err != nil{
			log.Error(err)
			return QError{key, "Could not load query associated with key"}
		}
		ctx.Queries[key] = q
	}
	return nil
}


// LoadQuery
// given the name of the query, search the config structure to find the filename and the bindvars array.
// Loads the text of the string into the Query structs.
// Populates the map Context.Queries
func (ctx *Context) LoadQuery(key string) (Query, error) {
	var q Query
	// Read the query from a file into a string
	pth := viper.GetString(fmt.Sprintf("features.%s.Filename", key))
	if len(pth) == 0 {
		return q, QError{key, "key has no associated filename"}
	}
	if !path.IsAbs(pth) {
		pth = path.Join(ctx.Querypath, pth)
	}
	fp, err := os.Open(pth)
	if err != nil{
		log.Errorf("%s: err", key)
		return q, err
	}
	b, err := ioutil.ReadAll(fp)
	if err != nil{
		log.Error(err)
		return q, nil
	}
	// Find the names of the query parameters (bindvars) from the configuration
	bndvars := viper.GetStringSlice(fmt.Sprintf("features.%s.bindvars", key))
	q = Query{pth, string(b), bndvars}
	return q, nil
}

// Query uses the key to find a query from the context and executes the query against the database
// the results come back as a *sql.Rows. The query parameters are passed as varargs argument to this function
// parameters to the query can be converted using the ctx.ArrangeBindVars(string, map[string]interface{}) function.
func (ctx *Context) Query(db sql.DB, key string, args...interface{})  (*sql.Rows, error) {
	var res *sql.Rows
	qry, ok := ctx.Queries[key]
	if !ok {
		err := QError{key, "no query with key found in Context"}
		return res, err
	}
	if qry.Text == "" {
		return res, QError{key, "q.Text is empty probably not properly read from file."}
	}
	res, err := db.Query(qry.Text, args...)
	if err != nil{
		log.Debug(qry.Text)
		return res, err
	}
	return res, err
}

// ArrangeBindVars: look up a query by key and then apply Query.ArrangeBindVars to that query.
func (ctx *Context) ArrangeBindVars(key string, values map[string]string) []interface{} {
	var args []interface{}
	q, ok := ctx.Queries[key]
	if !ok {
		return args
	}
	args = q.ArrangeBindVars(values)
	return args
}

