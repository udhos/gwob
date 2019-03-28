[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/gwob/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/udhos/gwob?status.svg)](http://godoc.org/github.com/udhos/gwob)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/gwob)](https://goreportcard.com/report/github.com/udhos/gwob)

# gwob
gwob - Pure Go Golang parser for Wavefront .OBJ 3D geometry file format

# Install

## Install with Go Modules (Go 1.11 or higher)

    git clone https://github.com/udhos/gwob
    cd gwob
    go install

## Install without Go Modules (Go before 1.11)

    go get github.com/udhos/gwob
    cd ~/go/src/github.com/udhos/gwob
    go install github.com/udhos/gwob

# Usage

Import the package in your Go program:

    import "github.com/udhos/gwob"

Example:

    // Error handling omitted for simplicity.

    import "github.com/udhos/gwob"

    inputObj, errOpen := os.Open("gopher.obj") // open OBJ file

    options := &gwob.ObjParserOptions{} // parser options

    o, errObj := gwob.NewObjFromReader(fileObj, bufio.NewReader(inputObj), options) // parse/load OBJ

    // Scan OBJ groups
    for _, g := range o.Groups {
        // ...
    }

# Example

Run the example:

    cd example
    go run main.go

See directory [example](example). 

# Documentation

See the [GoDoc](http://godoc.org/github.com/udhos/gwob) documentation.
