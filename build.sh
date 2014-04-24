#!/bin/bash
set -e
cd $(dirname $0)

DEST_OS=$1
dest_file="cmd2http"
if [ "$DEST_OS" == "windows" ];then
  export GOOS=windows 
  export GOARCH=386
  dest_file="cmd2http.exe"
fi
go get -u github.com/daviddengcn/go-ljson-conf
go get -u github.com/hidu/goutils
go get -u gopkg.in/cookieo9/resources-go.v2
go build -o $dest_file -ldflags "-s -w"  cmd2http.go 
echo $(date +"%Y%m%d.%H%M%S") >res/version
zip -r res.zip res
rm res/version
cat res.zip>> $dest_file
zip -A $dest_file
mkdir -p dest/
mv $dest_file dest/
rm res.zip
