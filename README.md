# featex

Feature Extraction as a Service!

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

Running it then should be as simple as:

```console
$ make
$ ./bin/featex
```

### Testing

``make test``

## Design

This package allows you to perform feature extraction step of a data analysis product
by running database queries as controled by a network service. 

### Problem to solve

Feature extraction takes a long time and typically requires executing database queries against
a large corpus of historical data. When near real time prediction applications are built,
these features must be available over a network commonly in the form of an Http API.

### Design Philosophy

We want a system for deploying production grade web services to serve up these features that is extensible
and configurable, while being easy to understand and debug. Since the feature extraction routines are
commonly expressed as database queries, in SQL or NoSQL languages, we built this software based on the idea
of mapping each set of features to a database query that has been constructed ahead of time.

These queries are then made available over a web service which accepts requests such as 
"Compute demographic features for people in this set," and returns the result.
There are two main types of requests, bulk extraction for building training and testing datasets, 
and single row extraction for serving up near real time prediction services.

### Configuration

This package uses spf13/viper to read configuration from ENV variables and TOML,YAML, or JSON files.
The configuration tells the system which routes map to which query files, and how to supply the 
bindvars (query parameters) to the DBMS.

### Query location

We accept a directory of queries such as `./sql` which contains SQL queries to be executed against a RDBMS such as Postgres.

## Config Specification

The configuration file is structured to map each key to a query. Each query must contain at least a filename and a list of bindvars.
An example configuration looks like:

```yaml
features:
  demographics:
    Filename: "demographics.sql"
    bindvars: ["limit"]
  features:
    Filename: "features.sql"
    bindvars: ["person"]
  demographics_historical:
    Filename: "demographics.historical.sql"
    bindvars: []
```

SQL queries can take positional query parameters, but URL query parameters and JSON, the lingua franca of the web,
yield keyword arguments in the web requests. The bindvars array is here to bridge that gap.
The bindvars array is list of the names of the positional query parameters in the order of their positions.
This tells the webserver how to arrange the keyword arguments that it receives in JSON or URL query parameters
into the order that the query expects for positional arguments.

# Conversation with Scott Brown
scott.brown [12:17 PM] 
I dunno

[12:17] 
I'm trying to write a simple end2end publish example

[12:17] 
where we train a model

[12:17] 
then publish

[12:18] 
then it's magically available on the website

[12:18] 
not sure how feature extraction fits into this

jpf 
[12:18 PM] 
ok then you don’t need to write new features

[12:18] 
you just need to call the api

scott.brown [12:18 PM] 
yah, what do I get from that?

[12:18] 
does it populate a table?

[12:18] 
to I get a sql query?

jpf 
[12:20 PM] 
use http://blackseabird01.icl.gtri.org:8080/queries

[12:20] 
to get the list of queries

[12:20] 
then you look at an entry like

```{"bulk_condition":
{
"Filename": "sql/bulk_condition.sql",
"Bindvars": ["cohort"],
"Text": "SELECT\n
 person_id,\n \n
 condition_era_start_date AS start_date,\n
 condition_era_end_date   AS end_date,\n \n
 condition_concept_id             concept_id,\n
 'condition'          AS type,\n
 1                        AS value\n
 FROM mimic_v5.condition_era\n\n
 -- cohort table : results_mimic.cohort\n
 RIGHT JOIN results_mimic.cohort on results_mimic.cohort.subject_id =
 person_id\n
 -- end\n\n
 WHERE cohort_definition_id= $1 and ( cohort_end_date \u003e condition_era_start_date ) 
 -- and not (condition_era_end_date \u003c cohort_start_date)\n
 ORDER BY person_id, type, start_date, end_date"
}
```

The key is the route, the Filename is the source file that contains the SQL, the bindvars are the names of the keyword arguments, and text is the text of the query


how would I run this bulk_condition query on a different dataset?

So you would then call that query with

curl http://blackseabird01.icl.gtri.org:8080/bulk/bulk_condition?cohort=121

jpf [12:24 PM] 
the service is tied to the one data set at startup
so we'd start a different service
for a different dataset
its tied to the connection string

[12:25] 
yeah when you run the service you pass an ENV variable with the database credentials

scott.brown [12:25 PM] 
I thought we were putting the different datasets in different postgres schemata

[12:26] 
how does that work with the connection string?

so you can configure that in the queries. There is a config file that has a SCHEMA variable to generate the queries with that schema.

It would be a different endpoint.
