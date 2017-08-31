#!/bin/sh

if [ $1 = "amd64" ]; then
    echo "x86_64"
elif [ $1 = "386" ]; then
    echo "i386"
else
    echo $1
fi
