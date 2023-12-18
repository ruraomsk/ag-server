#!/bin/bash
echo "Start to Linux deploy"
CGO_ENABLED=0 go build
if [ $? -ne 0 ]; then
	echo 'An error has occurred! Aborting the script execution...'
	exit 1
fi
# FILE=/home/rura/mnt/Linux/asud/cmd/ag-server
# if [ -f "$FILE" ]; then
#     echo "Mounted the server drive"
# else
#     echo "Mounting the server drive"
#     sudo mount.cifs -o username=root,password=162747 //192.168.115.23/asdu /home/rura/mnt/Linux
# fi
# sudo cp ag-server /home/rura/mnt/Linux/asud/cmd
