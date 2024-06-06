#!/bin/bash
echo "Start to Linux deploy"
CGO_ENABLED=0 go build
if [ $? -ne 0 ]; then
	echo 'An error has occurred! Aborting the script execution...'
	exit 1
fi
cp ag-server ~/tula/asud/cmd/