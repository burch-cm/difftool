package main

import (
	xxh "bitbucket.org/StephaneBunel/xxhash-go"
	"encoding/binary"
	"fmt"
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
	"github.com/deckarep/golang-set"
	"github.com/tealeg/xlsx"
	"log"
)

func buildPairs(xlfile string, indexcol int) map[string]uint32 {
	// open the xlsx file
	mySlice, err := xlsx.FileToSlice(xlfile)
	if err != nil {
		log.Fatalf("Unable to read file: %s\n", err)
	}
	// get dimensions of the file
	// ncol := len(mySlice[0][0])
	nrow := len(mySlice[0])
	// create map to hold output
	myMap := make(map[string]uint32)
	// iterate through file
	// start on f[1] to drop colnames
	for i := 1; i < nrow; i++ {
		barcode := mySlice[0][i][indexcol]
		hash := hashRow(mySlice[0][i])
		myMap[barcode] = hash
	}

	return myMap
}

/*
takes a slice of binaries and sums over them, returning an int32 value
*/
func byteSum(b []byte) int32 {
	var sum byte
	for _, v := range b {
		sum += v
	}
	return int32(sum)
}

/*
takes a slice of int32 values and sums them, returning a single int32 value
*/
func sum(x []int32) int32 {
	sum := x[0]
	for _, v := range x {
		sum += v
	}
	return sum
}

/*
takes a slice of string values from a line, converts each to a minary slice,
then performs a binary sum on that slice. Then each slice of binary sums is itself
summed. The integer sum of the binary sums is returned as a single uint32 value
*/
func hashRow(x []string) uint32 {
	var y []int32
	for _, v := range x {
		b := byteSum([]byte(v))
		y = append(y, b)
	}
	k := make([]byte, 8)
	binary.LittleEndian.PutUint32(k, uint32(sum(y)))
	return xxh.Checksum32(k)
}

/*convert interface to string*/
func conv2str(t interface{}) string {
	return fmt.Sprint(t)
}

/*
compares two maps by key value, and then checks to see if the overlapping keys
are mapped to equivalent hash values. determines if a key appears in left, right, or both maps.
*/
func compareHash(map1, map2 map[string]uint32) map[string]string {
	m1_key := mapset.NewSet()
	m2_key := mapset.NewSet()

	comparison := make(map[string]string)

	// iterate over key values
	for key := range map1 {
		m1_key.Add(key)
	}
	for key := range map2 {
		m2_key.Add(key)
	}
	// deleted keys - left difference
	left_only := m1_key.Difference(m2_key).ToSlice()
	for _, k := range left_only {
		comparison[conv2str(k)] = "removed"
	}
	// new keys - right difference
	right_only := m2_key.Difference(m1_key).ToSlice()
	for _, k := range right_only {
		comparison[conv2str(k)] = "new"
	}
	// shared keys - intersection of keys
	same_keys := m1_key.Intersect(m2_key).ToSlice()
	for _, k := range same_keys {
		if map1[conv2str(k)] == map2[conv2str(k)] {
			comparison[conv2str(k)] = "same"
		} else {
			comparison[conv2str(k)] = "different"
		}
	}

	return comparison
}

// write to a specified file
func writeToFile(m map[string]string, outfile string) {
	var file *xlsx.File
	var sheet *xlsx.Sheet
	// var row *xlsx.Row
	// var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}

	namerow := sheet.AddRow()
	cell := namerow.AddCell()
	cell.Value = "barcode"
	cell = namerow.AddCell()
	cell.Value = "difference"

	for k, v := range m {
		row1 := sheet.AddRow()
		cell1 := row1.AddCell()
		cell1.Value = k
		cell2 := row1.AddCell()
		cell2.Value = v
	}

	err = file.Save(outfile)
	if err != nil {
		fmt.Printf(err.Error())
	}
}

// knit it all together
func doTheThing(file1, file2, outfile string) bool {

	indexcol := 6
	m1 := buildPairs(file1, indexcol)
	m2 := buildPairs(file2, indexcol)

	writeToFile(compareHash(m1, m2), outfile)
	return true
}

// build the iterface
func setupUI() {
	mainwin := ui.NewWindow("File Select", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	hbox := ui.NewHorizontalBox()
	hbox.SetPadded(true)
	mainwin.SetChild(hbox)
	mainwin.SetMargined(true)

	grid := ui.NewGrid()
	grid.SetPadded(true)
	hbox.Append(grid, false)

	// Get File 1
	button := ui.NewButton("Open File")
	entry := ui.NewEntry()
	entry.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename1 := ui.OpenFile(mainwin)
		if filename1 == "" {
			filename1 = "(cancelled)"
		}
		entry.SetText(filename1)
	})
	grid.Append(button,
		0, 0, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry,
		1, 0, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	// Get File 2
	button = ui.NewButton("Open File")
	entry2 := ui.NewEntry()
	entry2.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename2 := ui.OpenFile(mainwin)
		if filename2 == "" {
			filename2 = "(cancelled)"
		}
		entry2.SetText(filename2)
	})
	grid.Append(button,
		0, 1, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry2,
		1, 1, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	// Get outfile
	button = ui.NewButton("Save File")
	entry3 := ui.NewEntry()
	entry3.SetReadOnly(true)
	button.OnClicked(func(*ui.Button) {
		filename3 := ui.SaveFile(mainwin)
		if filename3 == "" {
			filename3 = "(cancelled)"
		}
		entry3.SetText(filename3)
	})
	grid.Append(button,
		0, 2, 1, 1,
		false, ui.AlignFill, false, ui.AlignFill)
	grid.Append(entry3,
		1, 2, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	// Run the application
	button = ui.NewButton("Compare Files")
	button.OnClicked(func(*ui.Button) {
		if (entry.Text() == "") || (entry2.Text() == "") || (entry3.Text() == "") {
			ui.MsgBoxError(mainwin, "Error", "One or more filenames is invalid.")
			return
		}
		comp := doTheThing(entry.Text(), entry2.Text(), entry3.Text())
		if comp == true {
			msg := "Complete! Wrote result to " + entry3.Text()
			ui.MsgBox(mainwin, "Complete!", msg)
		} else {
			ui.MsgBoxError(mainwin, "Error", "Make sure all files are selected first.")
		}
	})

	grid.Append(button,
		0, 3, 1, 1,
		true, ui.AlignFill, false, ui.AlignFill)

	mainwin.Show()
}

func main() {
	ui.Main(setupUI)
}
