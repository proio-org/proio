#!/bin/bash

importfile=$1
protofile=$2

importstring=$(cat $protofile | grep 'option\s*go_package' | sed 's/option\s*go_package\s*=\s*"\(.*\)";/\1/' | sed 's/\//\\\//g')

if ! [ -f $importfile ]; then
	printf 'package proio\n\nimport (\n)\n' >> $importfile
fi

sed -i "s/)/\t_ \"$importstring\"\n)/" $importfile
