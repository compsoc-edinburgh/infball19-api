#! /bin/sh

echo "building infball_backend..."

rm -f infball_backend.tar.gz

docker build -t infball_backend .
