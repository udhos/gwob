#!/bin/bash

me=`basename $0`
msg() {
    echo >&2 $me: $*
}

msg fmt
gofmt -s -w *.go

msg fix
go tool fix *.go

msg vet
go tool vet .

msg install
go install

msg gosimple
hash gosimple 2>/dev/null && gosimple *.go

msg golint
hash golint 2>/dev/null && golint *.go

msg staticcheck
hash staticcheck 2>/dev/null && staticcheck *.go

msg test
go test

