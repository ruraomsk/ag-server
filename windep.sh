#!/bin/bash
echo "Start to Windows deploy"
GOOS=windows GOARCH=amd64 go build
FILE=/home/rura/mnt/ASDU/asud/cmd/ag-server.exe
if [ -f "$FILE" ]; then
    echo "Mounted the server drive"
else
    echo "Mounting the server drive"
    sudo mount -t cifs -o username=asdu,password=162747 \\\\192.168.115.115\\d /home/rura/mnt/ASDU
fi
sudo cp ./data/*.sql /home/rura/mnt/ASDU/asud/setup
sudo cp ./data/*.mrk /home/rura/mnt/ASDU/asud/setup
sudo cp ./data/*.xml /home/rura/mnt/ASDU/asud/setup
sudo cp ag-server.exe /home/rura/mnt/ASDU/asud/cmd
sudo cp *.toml /home/rura/mnt/ASDU/asud/cmd
sudo cp save.bat /home/rura/mnt/ASDU/asud/cmd

# cp ./data/*.sql ~/vm/asud/setup
# cp ./data/*.mrk ~/vm/asud/setup
# cp ./data/*.xml ~/vm/asud/setup
# cp ag-server.exe ~/vm/asud/cmd
# cp *.toml ~/vm/asud/cmd
