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
[ -x ~go/bin/gosimple ] && ~go/bin/gosimple *.go

msg golint
[ -x ~go/bin/golint ] && ~go/bin/golint *.go

msg staticcheck
[ -x ~go/bin/staticcheck ] && ~go/bin/staticcheck *.go

msg test
go test

