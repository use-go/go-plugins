#!/bin/bash -e

tag=$1

for m in $(find ./ -name 'go.mod'); do
  d=$(dirname $m);
  pushd $d;
  PKGS=$(grep github.com/micro/go-plugins go.mod | grep -v module | tr -s '\n' ' ' | awk '{print $1}')
  for PKG in $PKGS; do
    /bin/bash -c "go get $PKG/v2@$tag && go mod tidy" &
  done
  popd;
done
