#!/bin/bash
echo "Start to Windows deploy"
GOOS=windows GOARCH=amd64 go build
if [ $? -ne 0 ]; then
	echo 'An error has occurred! Aborting the script execution...'
	exit 1
fi
FILE=/home/rura/mnt/Debug/ASDU/asud/cmd/ag-server.exe
if [ -f "$FILE" ]; then
    echo "Mounted the server drive"
else
    echo "Mounting the server drive"
    sudo mount -t cifs -o username=rura,password=162747 \\\\192.168.115.25\\C /home/rura/mnt/Debug
fi
sudo cp ag-server.exe /home/rura/mnt/Debug/ASDU/asud/cmd
