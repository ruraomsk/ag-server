#!/bin/bash
echo "Start deploy"
cp ./setup/*.sql ~/setup
cp ./setup/*.json ~/setup
rm ~/log/ag-server/*.log
./ag-server
