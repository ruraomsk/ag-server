#!/bin/bash
echo "Start to Windows deploy"
GOOS=windows GOARCH=amd64 go build
FILE=/mnt/asud/cmd/ag-server.exe
if [ -f "$FILE" ]; then
    echo "Mounted the server drive"
else
    echo "Mounting the server drive"
    sudo mount -t cifs -o username=asdu,password=162747 \\\\192.168.115.115\\d /mnt
fi
sudo  cp ./data/*.sql /mnt/asud/setup
sudo  cp ./data/*.mrk /mnt/asud/setup
sudo  cp ./data/*.xml /mnt/asud/setup
sudo  cp ag-server.exe /mnt/asud/cmd
sudo  cp *.toml /mnt/asud/cmd
sudo  cp save.bat /mnt/asud/cmd

# cp ./data/*.sql ~/vm/asud/setup
# cp ./data/*.mrk ~/vm/asud/setup
# cp ./data/*.xml ~/vm/asud/setup
# cp ag-server.exe ~/vm/asud/cmd
# cp *.toml ~/vm/asud/cmd
