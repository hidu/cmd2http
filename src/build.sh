#!/bin/bash
set -e

go build cmd2http.go
zip -r res.zip res
cat res.zip>> cmd2http
zip -A cmd2http
mv cmd2http ../bin/
rm res.zip