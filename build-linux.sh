#!/bin/bash
docker build --pull --rm -f "Dockerfile" -t backupmysqlcli:latest "."
docker run -it -v ${PWD}:/app backupmysqlcli:latest