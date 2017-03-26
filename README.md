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
