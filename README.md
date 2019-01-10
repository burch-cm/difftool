# diffTool

## 
diffTool is a command-line application built in [Go](https://golang.org).

It finds the diff status of each keyed line between two Excel files by key (default is col 6).

## Interface  
The compiled file is called directly from the command line with flagged arguments:
difftool.exe -file1=filename1.xlsx -file2=filename2.xlsx -outfile=resultfilename.xlsx

This will produce an Excel .xlsx file with two columns - key and status. Key is the key value of the keyed column used for the set analysis between files, and status takes one of four values: new, removed, changed, same

## How it works
...
