#!/bin/sh
LOC=$(dirname "$0")
killall main

# Delete old logs.
find $LOC/../logs -mindepth 1 -type f -mtime +2 -delete
export GOOGLE_APPLICATION_CREDENTIALS=$LOC/../resources/mindfront-tech-0fd4ed107e8d.json

$LOC/main \
				-log_dir=$LOC/../logs/ \
				-resources_path=$LOC/../resources \
				-alsologtostderr=false \
				-logtostderr=false \
				-en_emotion=true \
				-v=2
