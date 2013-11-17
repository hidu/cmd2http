#!/bin/bash
set -e

go build -ldflags "-s -w" cmd2http.go
echo "1.0 " $(date +"%Y%m%d.%H%M%S") >res/version
zip -r res.zip res
rm res/version
cat res.zip>> cmd2http
zip -A cmd2http
mv cmd2http ../dest/
rm res.zip