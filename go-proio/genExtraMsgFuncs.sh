#!/bin/bash

protoFile=$1
outFile=$2

collections=$(grep "^message.*Collection\>" $protoFile | sed "s/^.*\<\(\S*Collection\)\>.*/\1/")
messages=$(grep "^message.*Collection\>" $protoFile | sed "s/^.*\<\(\S*\)Collection\>.*/\1/")

printf "\n\n// Extra generated functions for compliance with Message and Collection interfaces\n\n" >> $outFile

for coll in $collections; do
	printf "func (c *$coll) SetId(id uint32) {\n\
		c.Id = id\n\
	}\n\n" >> $outFile

	printf "func (c *$coll) GetNEntries() uint32 {\n\
		return uint32(len(c.Entries))\n\
	}\n\n" >> $outFile

	printf "func (c *$coll) GetEntry(i uint32) proto.Message {\n\
		if i < uint32(len(c.Entries)) {\n\
			return c.Entries[i]\n\
		}
		return nil
	}\n\n" >> $outFile
done

for msg in $messages; do
	printf "func (m *$msg) SetId(id uint32) {\n\
		m.Id = id\n\
	}\n\n" >> $outFile
done

gofmt -w $outFile
