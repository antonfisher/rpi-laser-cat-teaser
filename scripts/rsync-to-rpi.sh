#!/usr/bin/env bash

make
rsync -avz ./bin/ pi@10.0.0.82:/home/pi/app;

while inotifywait -r -e modify,create,delete ./cmd ./pkg; do
    make
    rsync -avz ./bin/ pi@10.0.0.82:/home/pi/app;
done;
