#!/bin/bash

me=$(basename "$0")
msg() {
    echo >&2 "$me:" "$@"
}

# this will acidentally install shadow as a dependency 
hash shadow 2>/dev/null || go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow

gofmt -s -w ./*.go ./example
go tool fix ./*.go ./example
go vet -vettool="$(which shadow)" . ./example
go install

#hash gosimple 2>/dev/null    && gosimple    ./*.go
hash golint 2>/dev/null      && golint      ./*.go
#hash staticcheck 2>/dev/null && staticcheck ./*.go

#hash gosimple 2>/dev/null    && gosimple    ./example/*.go
hash golint 2>/dev/null      && golint      ./example/*.go
#hash staticcheck 2>/dev/null && staticcheck ./example/*.go

go test
go test -bench=.

go mod tidy ;# remove non-required modules from dependencies
