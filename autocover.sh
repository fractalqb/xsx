#!/bin/sh
#WATCH=
while inotifywait -e move_self -e modify *.go; do
    make
done
