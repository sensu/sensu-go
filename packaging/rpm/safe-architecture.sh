#!/bin/sh

if [ $1 = "amd64" ]; then
    echo "x86_64"
elif [ $1 = "i386" ]; then
    echo "x86"
else
    echo $1
fi
