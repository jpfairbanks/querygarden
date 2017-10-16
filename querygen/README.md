# QUERYGEN

querygen is a package for generating sql automatically from templates.
This package is meant to statically generate all the sql that a program will
need at deployment/compile time rather than at run time.

This program solves a problem that many analytic software packages have.
We have a set of databases that have similar schemas, but small variations in the
particular configuration of the database. The variations are in areas such as table 
names, which means that the variation cannot be captured in query parameters.
Thus we need to generate the sql programatically. 
The simplest technique for this is to just concatenate strings together, 
which leads to repetitive code, and opens the system to SQL injection vulnerabilities.
By using Object Relational Mapping (ORM), one can avoid the drawbacks of manually creating SQL strings.
However, ORM systems are designed for CRUD applications and are tied to the programming language of the ORM library.

querygen attempts to solve this problem by generating all the SQL queries ahead of time.
These statically generated queries can be inspected by the user, explained by the database, and shared accross multiple 
applications.
This allows the user to audit the generated queries for security vulnerabilities knowing that the only run time manipulation
of these queries will use query parameters and the database driver's escaping.


## Installation

go get github.com/jpfairbanks/querygen

## Usage

You first need to make a directory of go text/template templates.
By default this directory should be called templates.
Each template is a file that looks like 

For example templates/demographics.sql.tmpl:
```mustache
select gender, race, age, count(*) from {{.schema}}.{{.table}} where
  {{.cond}}
group by
  gender, race, age
{{.limit}}
```

Then you write a configuration file to tell querygen how to render your templates.
The config file can be written in any syntax supported by [viper](github.com/spf13/viper).

```yaml
global:
  schema: "SCHEMA"
  version: "v0.1"
scopes:
  - demographics
  - demographics_historical
demographics:
  Template: "demographics.sql.tmpl"
  Filename: "demographics.sql"
  Limit: "LIMIT ?"
  Cond: "year <= 2016 and year >= 2000"
  Table: "table"
demographics_historical:
  Template: "demographics.sql.tmpl"
  Filename: "demographics.historical.sql"
  Limit: ""
  Cond: "year <= 2000 and year >= 1900"
  Table: "table"%
```
The config has a set of global variables associated with the key "global".
These global variables will be made available to all templates.
There is a list of scopes, which tells the program which templates to render.
Each scope contains a set of variable that will be available to the template when running.
Two special variables, Template and Filename control the behavior of the program.
The Template variable is a path to find the template in a file.
The Filename variable is a path to store the rendered template.
The reason for these two variable is to allow the same template to be used with
multiple scopes.


Then you run the script. Each template will be rendered into the file specified in the config file.
```bash
./querygen 
```

Based on the example above, the output would be 

```postgresql
select gender, race, age, count(*) from SCHEMA.table where
  year <= 2016 and year >= 2000
group by
  gender, race, age
LIMIT ?
```

## Future Work

- [ ] Expand to arbitrary templating frameworks with golang implementations