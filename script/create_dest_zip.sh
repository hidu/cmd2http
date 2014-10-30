#!/bin/bash
cd $(dirname $0)
cd ../

if [ -z "$1" ];then
    bash build.sh
    bash build.sh windows
fi
NAME=cmd2http

VERSION=$(cat res/version)

cd dest
################################################

if [ -d target ];then
  rm -rf target
fi
mkdir -p target/conf/
mkdir -p target/bin/

cp ../example/cmd2http.conf target/conf/
cp ../script/cmd2http_control.sh ../script/windows_run.bat  target/bin/
cp ../example/cmds/ target/conf -a
cp ../example/data/ target/conf -a

cp $NAME $NAME.exe target/bin/

t=$(date +"%Y%m%d")

rm "${NAME}_"* -rf

NEW_NAME="${NAME}_${VERSION}"
mv target "$NEW_NAME"

################################################
TAR_NAME="${NAME}_${VERSION}_$t.tar.gz"

tar -czvf "$TAR_NAME" "$NEW_NAME"
