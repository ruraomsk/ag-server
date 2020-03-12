#!/bin/bash
echo "Start to Windows deploy"
GOOS=windows GOARCH=amd64 go build
cp ./data/*.sql ~/vm/asud/setup
cp ./data/*.mrk ~/vm/asud/setup
cp ./data/*.xml ~/vm/asud/setup
cp ag-server.exe ~/vm/asud/cmd
cp *.toml ~/vm/asud/cmd
