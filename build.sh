#!/bin/bash
set -e
cd $(dirname $0)

go build -ldflags "-s -w" cmd2http.go
echo $(date +"%Y%m%d.%H%M%S") >res/version
zip -r res.zip res
rm res/version
cat res.zip>> cmd2http
zip -A cmd2http
mv cmd2http dest/
rm res.zip