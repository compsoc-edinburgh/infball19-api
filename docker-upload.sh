#! /bin/bash
# poor man's kubernetes

set -e

#./docker-build.sh

printf "copying infball_backend to "
echo $1

docker save -o infball_backend.tar.gz infball_backend
scp infball_backend.tar.gz $1:~


