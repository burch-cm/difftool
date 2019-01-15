# diffTool

## 
diffTool is a GUI application built in [Go](https://golang.org). It finds the diff status of each keyed line between two Excel files by key and then writes the differences to a specificed file.


## Installation
To run diffTool on a Windows-based system, download a copy of diffTool.exe and run it from your computer. To build the diffTool on any go-capable system, clone the repo and then run 
``` 
go build diffTool.go
```
from within the repo folder.

## Use  

Please read the instructions [here](https://burch-cm.github.io/difftool/)

This will produce an Excel .xlsx file with two columns - key and status. Key is the key value of the keyed column used for the set analysis between files, and status takes one of four values: new, removed, changed, same

## How it works  
Under the hood, diffTool reads each file into a Go map using the selected key column and a hash of the line containing that key. Each set of keys is compared for differences and intersections, and the intersectional keys have their line hash values compared to check for differences. The results are paired back to the keys in a key: value map, and then written to a specified file.
