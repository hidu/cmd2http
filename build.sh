#!/bin/bash

# build for window: build.sh windows
# default linux

set -e
cd $(dirname $0)

DEST_OS=$1
dest_file="cmd2http"
if [ "$DEST_OS" == "windows" ];then
  export GOOS=windows 
  export GOARCH=386
  dest_file="cmd2http.exe"
fi
go build -o $dest_file -ldflags "-s -w"  cmd2http.go 
zip -r res.zip res
cat res.zip>> $dest_file
zip -A $dest_file
mkdir -p dest/
mv $dest_file dest/
rm res.zip

echo "done"