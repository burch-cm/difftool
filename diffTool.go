package main

import (
	"fmt"
	"github.com/burch-cm/xlsx"
	xxh "github.com/burch-cm/xxhash"
	"github.com/deckarep/golang-set"
	"log"
	"regexp"
	"strings"
)

// Type rowMap holds a slice of columns values as strings mapped to a key as string.
type rowMap map[string][]string
type diffMap map[string]map[string][]string

// Type roeMapKey holds the set difference and intersection of the keys from two rowMaps
type rowMapKey struct {
	removed      []string
	added        []string
	intersection []string
	different    []string
}

func (r rowMap) Lower() {
	for key, val := range r {
		var lower []string
		for _, j := range val {
			lower = append(lower, strings.ToLower(j))
		}
		r[key] = lower
	}
}

// Method colHash applies the xxhash algorithm to each value in a rowMap column slice and returns those values.
func (r rowMap) colHash(key string) []uint64 {
	t := r[key]
	var out []uint64
	for _, v := range t {
		out = append(out, xxh.Sum64String(v))
	}
	return out
}

// Method rowHash applies the xxhash algorithm to each value in a rowMap column slice and returns the sum.
func (r rowMap) rowHash(key string) uint64 {
	vals := r.colHash(key)
	var sum uint64 = 0
	for _, v := range vals {
		sum += v
	}
	return sum
}

func conv2str(t interface{}) string {
	return fmt.Sprint(t)
}

func getColNames(xlfile string) []string {
	mySlice, err := xlsx.FileToSliceNlines(xlfile, 1)
	if err != nil {
		log.Fatalf("Unable to open file: %s\n", err)
	}
	return mySlice[0][0]
}

//RowMap reads in an excel file and returns a rowMap object and a string slice of column names.
func RowMap(xlfile string, indexpos int) (r rowMap, colnames []string) {
	mySlice, err := xlsx.FileToSlice(xlfile)
	if err != nil {
		log.Fatalf("Unable to read file: %s\n", err)
	}
	ncol := len(mySlice[0][0])
	nrow := len(mySlice[0])
	var colNames []string
	for i := 0; i < ncol; i++ {
		colNames = append(colNames, (mySlice[0][0][i]))
	}
	var myMap = make(rowMap)
	for i := 1; i < nrow; i++ {
		keyval := mySlice[0][i][indexpos]
		myMap[keyval] = mySlice[0][i]
	}
	myMap.Lower()
	return myMap, colNames
}

func CompKeys(m1, m2 rowMap) rowMapKey {
	m1_key := mapset.NewSet()
	m2_key := mapset.NewSet()
	var out rowMapKey
	for key := range m1 {
		m1_key.Add(key)
	}
	for key := range m2 {
		m2_key.Add(key)
	}
	rem := m1_key.Difference(m2_key).ToSlice()
	for _, i := range rem {
		out.removed = append(out.removed, conv2str(i))
	}
	rem = m2_key.Difference(m1_key).ToSlice()
	for _, i := range rem {
		out.added = append(out.added, conv2str(i))
	}
	rem = m1_key.Intersect(m2_key).ToSlice()
	for _, i := range rem {
		out.intersection = append(out.intersection, conv2str(i))
	}
	/*
		out.removed = m1_key.Difference(m2_key).ToSlice()
		out.added = m2_key.Difference(m1_key).ToSlice()
		out.intersection = m1_key.Intersect(m2_key).ToSlice()
	*/
	return out
}

func sameHash(m1, m2 rowMap, keyval string) bool {
	if m1.rowHash(keyval) == m2.rowHash(keyval) {
		return true
	}
	return false
}

func Difference(xlfile1, xlfile2 string, indexpos int) (diffMap, []string) {
	rm1, colnames1 := RowMap(xlfile1, indexpos)
	rm2, _ := RowMap(xlfile2, indexpos)
	keyset := CompKeys(rm1, rm2)
	for _, v := range keyset.intersection {
		if sameHash(rm1, rm2, v) == false {
			keyset.different = append(keyset.different, v)
		}
	}

	var outMap = make(diffMap)
	outMap["colnames"] = make(map[string][]string)
	outMap["colnames"]["colnames"] = colnames1
	outMap["colnames"]["type"] = []string{"colnames"}

	for _, v := range keyset.different {
		outMap[v] = make(map[string][]string)
		outMap[v]["type"] = []string{"different"}
		outMap[v]["old"] = rm1[v]
		outMap[v]["new"] = rm2[v]
	}
	for _, v := range keyset.added {
		outMap[v] = make(map[string][]string)
		outMap[v]["type"] = []string{"added"}
		outMap[v]["new"] = rm2[v]
	}
	for _, v := range keyset.removed {
		outMap[v] = make(map[string][]string)
		outMap[v]["type"] = []string{"removed"}
		outMap[v]["old"] = rm1[v]
	}
	return outMap, colnames1
}

// as a method for a diffMap
func (diff diffMap) writeFile(f string) bool {
	matched, err := regexp.MatchString(".xlsx$|.csv$", f)
	if matched != true {
		f = f + ".xlsx"
	}
	var file *xlsx.File
	var sheet *xlsx.Sheet
	style := xlsx.NewStyle()
	myFont := xlsx.NewFont(11, "Calibri")
	style.Font = *myFont

	file = xlsx.NewFile()
	// Differences
	sheet, err = file.AddSheet("Differences")
	if err != nil {
		fmt.Println(err.Error())
	}

	namerow := sheet.AddRow()
	cell := namerow.AddCell()
	cell.Value = "Row Source"
	cell.SetStyle(style)
	for _, v := range diff["colnames"]["colnames"] {
		cell := namerow.AddCell()
		cell.Value = v
		cell.SetStyle(style)
	}

	for _, v := range diff {
		if v["type"][0] == "different" {
			oldrow := sheet.AddRow()
			cell := oldrow.AddCell()
			cell.Value = "file 1"
			cell.SetStyle(style)
			for _, j := range v["old"] {
				cell := oldrow.AddCell()
				cell.Value = j
				cell.SetStyle(style)
			}
			newrow := sheet.AddRow()
			cell = newrow.AddCell()
			cell.Value = "file 2"
			cell.SetStyle(style)
			for _, j := range v["new"] {
				cell := newrow.AddCell()
				cell.Value = j
				cell.SetStyle(style)
			}
		}
	}
	// end Differences
	sheet, err = file.AddSheet("Added")
	if err != nil {
		fmt.Println(err.Error())
	}

	namerow = sheet.AddRow()
	for _, v := range diff["colnames"]["colnames"] {
		cell := namerow.AddCell()
		cell.Value = v
		cell.SetStyle(style)
	}
	for _, v := range diff {
		if v["type"][0] == "added" {
			oldrow := sheet.AddRow()
			for _, j := range v["old"] {
				cell := oldrow.AddCell()
				cell.Value = j
				cell.SetStyle(style)
			}
		}
	}
	// end Additions
	sheet, err = file.AddSheet("Removed")
	if err != nil {
		fmt.Println(err.Error())
	}

	namerow = sheet.AddRow()
	for _, v := range diff["colnames"]["colnames"] {
		cell := namerow.AddCell()
		cell.Value = v
		cell.SetStyle(style)
	}
	for _, v := range diff {
		if v["type"][0] == "removed" {
			oldrow := sheet.AddRow()
			for _, j := range v["old"] {
				cell := oldrow.AddCell()
				cell.Value = j
				cell.SetStyle(style)
			}
		}
	}
	// end Removed

	err = file.Save(f)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else {
		return true
	}

}

/*func main() {
	f1 := "C:/users/chris/Documents/GLT/working/AITS/AITS_Datadump_1_short.xlsx"
	f2 := "C:/users/chris/Documents/GLT/working/AITS/AITS_Datadump_2_short.xlsx"
	f1 := "C:/users/chris/Documents/GLT/working/AITS/AITS_Datadump_1023201.xlsx"
	f2 := "C:/users/chris/Documents/GLT/working/AITS/AITS_Datadump_11132018.xlsx"
	res := Difference(f1, f2, 6)

	res.writeFile("./test/testfile_method.xlsx")
	//writeFile(res, "./test/testfile.xlsx")
}*/
