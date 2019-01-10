# diffTool

## 
diffTool is a command-line application built in [Go](https://golang.org).

It finds the diff status of each keyed line between two Excel files by key.

## Interface  
The compiled file is called directly from the command line with flagged arguments:  
```
difftool.exe -file1=filename1.xlsx -file2=filename2.xlsx -outfile=resultfilename.xlsx  
```
This will produce an Excel .xlsx file with two columns - key and status. Key is the key value of the keyed column used for the set analysis between files, and status takes one of four values: new, removed, changed, same

## How it works  
Under the hood, diffTool reads each file into a Go map using the selected key column and a hash of the line containing that key. Each set of keys is compared for differences and intersections, and the intersectional keys have their line hash values compared to check for differences. The results are paired back to the keys in a key: value map, and then written to a specified file.
