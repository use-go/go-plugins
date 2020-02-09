#!/bin/bash -ex

mod=$(go list -m | sed 's|/v2||g')
PKGS=""
for d in $(find * -name 'go.mod'); do
  PKGS=" $PKGS ${mod}/$(dirname $d)/v2"
done

go mod download $PKGS
