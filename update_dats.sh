#!/bin/bash

datzip=`curl -s https://download.nai.com/products/commonupdater/current/vscandat1000/dat/0000/ | egrep -m1 -o "\"avvdat-[0-9]+\.zip\"" | cut -d '"' -f 2`

#download dat file
curl -o /tmp/$datzip -s "https://download.nai.com/products/commonupdater/current/vscandat1000/dat/0000/$datzip"

#install dat files
unzip -o -d /usr/local/uvscan /tmp/$datzip

#clean up
rm -rf /tmp/$datzip

#decompresss new dats
/usr/local/bin/uvscan --decompress /usr/local/uvscan/ 
