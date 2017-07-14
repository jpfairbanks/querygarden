# featex

Feature Extraction as a Service!

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

Running it then should be as simple as:

Dependencies are handled by glide so you will need to `go get github.com/Masterminds/glide` in order to download the dependencies and vendor them.


```console
$ make get-deps
$ go generate
$ make build 
$ ./bin/featex
```

You will need to edit the templates and configuration files before this repo
does anything useful for you. So you should lean about the sqlgen package before
using this package. The templates can be found in the `./templates` directory.
The contain the queries that this service will execute.

### Testing

``make test``

## Design

This package allows you to perform feature extraction step of a data analysis product
by running database queries over a network service.

### Problem to solve

Feature extraction takes a long time and typically requires executing database queries against a
large corpus of historical data. When near real time prediction applications are built, these
features must be available over a network commonly in the form of an HTTP API.

FeatEx supports both forms a long running `bulk` API and a single record `query` API.

### Design Philosophy

We want a system for deploying production grade web services to serve up these features that is extensible
and configurable, while being easy to understand and debug. Since the feature extraction routines are
commonly expressed as database queries in SQL or NoSQL languages, we built this software based on the idea
of mapping each set of features to a database query that has been constructed ahead of time. Ahead
of time construction of the queries is chosen to prevent SQL injection attacks and enable
administrators to inspect the queries offline.

These queries are then made available over a web service which accepts requests such as
"Compute demographic features for people in this set," and returns the result.
There are two main types of requests, bulk extraction for building training and testing datasets,
and single row extraction for serving up near real time prediction services.

### Configuration

This package uses spf13/viper to read configuration from ENV variables and TOML,YAML, or JSON files.
The configuration tells the system which routes map to which query files, and how to supply the
bindvars (query parameters) to the DBMS.

#### Database connection information

In order to tell featex how to connect to the database you should export an
enviroment variable DBSTRING.

`export DBSTRING="postgres://user:passwd@host:5432/postgres"`

You may need to disable sslmode for postgres although this indicates you should
review the security posture of your database.
`export DBSTRING="postgres://user:passwd@host:5432/postgres?sslmode=disable"`

Then run the binary and it will have credentials for connecting to the database.

#### Query location

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

## Endpoints

### Queries

You can get the set of available queries at the endpoint `/queries`. This url
returns the internal representation of the queries and can be used for
debugging.

The structure of these fields are:

```json
"queryname": {
"Filename": "sql/queryname.sql",
"Text": "select subject_id as person_id, cohort_start_date as start_date, cohort_end_date as end_date, cohort_definition_id as concept_id, 'cohort_id' as type, 1 as value from results_mimic.cohort\nwhere cohort_definition_id = $1\n",
"Bindvars": ["cohort"]}
```
- key: the name of the query.
- Filename: is a path on the server where the query can be found.
- Text: is the actual text of the SQL query
- Bindvars: provides keyword arguments for the query in the order that they are positional.

Here is an example of the output.
```json
{
"cohortpatients": {
"Filename": "sql/cohortpatients.sql",
"Text": "select subject_id as person_id, cohort_start_date as start_date, cohort_end_date as end_date, cohort_definition_id as concept_id, 'cohort_id' as type, 1 as value from results_mimic.cohort\nwhere cohort_definition_id = $1\n",
"Bindvars": [
"cohort"
]
},
"condition": {
"Filename": "sql/condition.sql",
"Text": "SELECT\n  person_id,\n  condition_start_date AS start_date,\n  condition_end_date   AS end_date,\n  condition_concept_id    concept_id,\n  'condition'          AS type,\n  1                    AS value\nFROM mimic_v5.condition_occurrence\nWHERE person_id = $1\nORDER BY person_id, type, start_date, end_date",
"Bindvars": [
"person"
]
},
"demographics": {
"Filename": "sql/demographics.sql",
"Text": "select gender, race, age, count(*) from mimic_v5.table where\n  year <= 2016 and year >= 2000\ngroupby\n  gender, race, age\nLIMIT ?",
"Bindvars": [
"limit"
]
}
"bulk_condition": {
"Filename": "sql/bulk_condition.sql",
"Text": "SELECT\n person_id,\n \n  condition_era_start_date AS start_date,\n  condition_era_end_date   AS end_date,\n \n condition_concept_id             concept_id,\n 'condition'          AS type,\n 1                        AS value\nFROM mimic_v5.condition_era\n\n   -- cohort table : results_mimic.cohort\n  RIGHT JOIN results_mimic.cohort on results_mimic.cohort.subject_id = person_id\n  -- end\n\nWHERE cohort_definition_id= $1 and ( cohort_end_date > condition_era_start_date ) -- and not (condition_era_end_date < cohort_start_date)\nORDER BY person_id, type, start_date, end_date",
"Bindvars": [
"cohort"
]
},
"bulk_demographics": {
"Filename": "sql/bulk_demographics.sql",
"Text": "\n\n-- AGE as binned by OHDSI/FeatureExtraction\nselect person_id,\n        start_date,\n        end_date,\n        0 as concept_id, \n        age as value,\n        'age' as type\n  from  (\n  SELECT\n    person_id,\n    cohort_start_date                                    AS start_date,\n    cohort_end_date                                      AS end_date,\n    date_part('year', cohort_start_date) - year_of_birth AS age\n  FROM mimic_v5.person\n    RIGHT JOIN results_mimic.cohort ON subject_id = person_id\n  WHERE cohort_definition_id = $1\n) as t\n\n\nUNION ALL\n\n-- gender\nselect person_id, cohort_start_date as start_date, cohort_end_date as end_date,\n  gender_concept_id as concept_id, 1 as value, 'gender' as type\nfrom mimic_v5.person\n  RIGHT JOIN results_mimic.cohort on subject_id=person_id WHERE cohort_definition_id = $1\n\n\n\n\nUNION ALL\n-- race\nselect person_id, cohort_start_date as start_date, cohort_end_date as end_date,\n                  race_concept_id as concept_id, 1 as value, 'race' as type\nfrom mimic_v5.person\n  RIGHT JOIN results_mimic.cohort on subject_id=person_id WHERE cohort_definition_id = $1\n\n\n\n\n",
"Bindvars": [
"cohort"
]
},
}
```

### /Query/

You can access a query using the following endpoint.
`/query/<queryname>?key0=value0&key1=value1`
The all information necessary to identify and run the query is found after the prefix `/query/`.

The key-value pairs are used to identify the arguments to the query. They are specified in the
bindvars parameter of the object returned by `/queries/`. If there are required arguments missing,
you will get an error message from the endpoint. We try to stick to informative HTTP error codes
rather than implementing a JSON based exception system.

# Known Issues / TODO

- Multi-tenancy: currently this project only supports one database at a time and
  one set of queries.

  Full multi-tenancy would require a user and permissions model where users
    have access to queries, and queries can be run on databases. Where the
    authentication and authorization rests on a triple (user, query, database).
    This would allow administrators to control what queries get run on
    their databases and which users can see the results of which queries.

- Dynamic queries: part of the design of this service is to use only static
  queries to minimize the opportunity for SQL injection attacks. If we are
  opening up a service to the network to allow people to run arbitrary database
  queries, it will be hard to protect against SQL injection. It would be
  possible to build a dynamic query engine, but that really opens up the attack
  surface. The current model is make it really easy to add statically generated
  queries using sqlgen and some yaml files. Anything that can go in an SQL
  parameterized query is handled by the the DB library. Anything else needs to
  be statically rendered to disk ahead of time. This allows the administrator to
  audit the set of available queries and make sure there is no risk of injection
  attacks.

- External persistence of the queries: if we allow users to upload queries to
  this services, it will be important to persist them into a database. Given the
  current design of the package it would be easy to store the queries in a key
  value store that supported JSON or BSON values and hierarchical keys like
  file paths.
