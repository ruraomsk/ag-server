#!/bin/bash
echo "Start deploy"
cp ./data/*.sql ~/setup
cp ./data/*.json ~/setup
cp ./data/*.mrk ~/setup
cp ./data/*.xml ~/setup
rm ~/log/ag-server/*.log
./ag-server %1
