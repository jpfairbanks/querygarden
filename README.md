# QueryGarden

Growing applications in fertile SQL.

This repository holds three sub projects. The main application is `featex` a feature extraction as a service application.
In order to build this application, there are two reusable projects that are bundled as dependencies.
The `garden` module which contains all the code for managing collections of queries for service over web protocols,
and the `querygen` program, which is a command line tool for generating queries from template specifications.

See [querygen/README.md]() and [featex/README.md]() for more.

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

Dependencies are handled by glide so you will need to `go get github.com/Masterminds/glide` in order to download the dependencies and vendor them.
The dependencies can also be installed manually by reading the glide file and installing them globally.

You can build the featex application using the makefile like so:

```bash
$ make build
building featex 0.1.0
GOPATH=/Users/jfairbanks6/golang/
cd featex && go build -ldflags "-X main.GitCommit=be2e20c0b4b4dc9650cf0c0b84d036e8d13c5285+CHANGES -X main.VersionPrerelease=DEV" -o bin/featex

$ cd featex && ./bin/featex
INFO[0000] resultsSchema=results_mimic
INFO[0000] Loading queries
INFO[0000] The keys are: [drugs milenial_features bulk_drugs demographics bulk_condition bulk_procedure condition demographics_historical cohortpatients drug_era bulk_demographics features]
INFO[0000] ./sql
INFO[0000] Database connected
INFO[0000] Read Templates                                Templates="; defined templates are: "queries.html.tmpl", "query.html.tmpl", "index.html.tmpl", "404.html.tmpl""
INFO[0000] Serving on address: :8080
```

See [featex/README.md]() for more details about the featex application.
