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

var (
    mainwin   *ui.Window
    fileNames []string
    indexcol  int = 6
)

func which(s []string, tgt string) (bool, int) {
    for i := 0; i < len(s); i++ {
        if s[i] == tgt {
            return true, i
        }
    }
    return false, 0
}

func getColNames(xlfile string) []string {
    mySlice, err := xlsx.FileToSlice(xlfile)
    if err != nil {
        log.Fatalf("Unable to read file: %s\n", err)
    }
    ncol := len(mySlice[0][0])
    var colNames []string
    for i := 1; i < ncol; i++ {
        colNames = append(colNames, (mySlice[0][0][i]))
    }
    return colNames
}

func updateColList(box *ui.Combobox, s []string) {
    for _, col := range s {
        box.Append(col)
    }
    ok, ind := which(s, "Barcode")
    if ok {
        box.SetSelected(ind)
    } else {
        box.SetSelected(0)
    }
}

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
/*func compareHash(map1, map2 map[string]uint32) map[string]string {
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
}*/

type HashComp struct {
    keys      map[string]string
    same      []string
    added     []string
    removed   []string
    different []string
}

func compareHash(map1, map2 map[string]uint32) map[string]string {
    m1_key := mapset.NewSet()
    m2_key := mapset.NewSet()
    comparison := make(map[string]string)
    same, diff, added, removed := []string{}

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
        removed = append(removed, conv2str(k))
    }
    // new keys - right difference
    right_only := m2_key.Difference(m1_key).ToSlice()
    for _, k := range right_only {
        comparison[conv2str(k)] = "added"
        added = append(added, conv2str(k))
    }
    // shared keys - intersection of keys
    same_keys := m1_key.Intersect(m2_key).ToSlice()
    for _, k := range same_keys {
        if map1[conv2str(k)] == map2[conv2str(k)] {
            comparison[conv2str(k)] = "same"
            same = append(same, conv2str(k))
        } else {
            comparison[conv2str(k)] = "different"
            different = append(different, k)
        }
    }
    out := HashComp{
        keys:      comparison,
        same:      same,
        different: different,
        added:     added,
        removed:   removed,
    }

    return out
}

// write to a specified file
/*func writeToFile(m map[string]string, outfile string) {
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
    cell.Value = "key"
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
*/
func writeToFile(c HashComp, outfile string) {
    var file *xlsx.File
    var sheet *xlsx.Sheet
    var err error

    file = xlsx.NewFile()
    sheet, err = file.AddSheet("Sheet1")
    if err != nil {
        fmt.Printf(err.Error())
    }

    namerow := sheet.AddRow()
    cell := namerow.AddCell()
    cell.Value = "key"
    cell = namerow.AddCell()
    cell.Value = "difference"

    /*    for k, v := range c {
              row1 := sheet.AddRow()
              cell1 := row1.AddCell()
              cell1.Value = k
              cell2 := row1.AddCell()
              cell2.Value = v
          }
    */
    for _, d := range c.different {
        row1 := sheet.AddRow()

    }

    err = file.Save(outfile)
    if err != nil {
        fmt.Printf(err.Error())
    }
}

// knit it all together
func doTheThing(file1, file2, outfile string, indexcol int) bool {

    m1 := buildPairs(file1, indexcol)
    m2 := buildPairs(file2, indexcol)
    hashcomp := compareHash(m1, m2)

    return true
}

// progress bar update
func withProgress(pb *ui.ProgressBar, entry1, entry2, entry3 *ui.Entry, indexcol int) bool {
    pb.SetValue(-1)
    outval := false
    if (entry1.Text() == "") || (entry2.Text() == "") || (entry3.Text() == "") {
        ui.MsgBoxError(mainwin, "File Select Error", "Please select input and output files first.")
        return outval
    }
    outval = doTheThing(entry1.Text(), entry2.Text(), entry3.Text(), indexcol)
    return outval

}

// input tab
func makeControlsPage() ui.Control {
    fileCols := ui.NewCombobox()
    hbox := ui.NewHorizontalBox()
    hbox.SetPadded(true)

    vbox := ui.NewVerticalBox()
    vbox.SetPadded(true)
    hbox.Append(vbox, true)

    // input grid
    grid := ui.NewGrid()
    grid.SetPadded(true)
    vbox.Append(grid, false)

    // file 1
    button := ui.NewButton("Select File 1")
    entry1 := ui.NewEntry()
    entry1.SetReadOnly(true)
    button.OnClicked(func(*ui.Button) {
        filename1 := ui.OpenFile(mainwin)
        if filename1 == "" {
            filename1 = ""
        }
        entry1.SetText(filename1)
        colnames := getColNames(filename1)
        updateColList(fileCols, colnames)
    })

    grid.Append(button,
        0, 0, 1, 1,
        false, ui.AlignFill, false, ui.AlignFill)
    grid.Append(entry1,
        1, 0, 1, 1,
        true, ui.AlignFill, false, ui.AlignFill)

    // file 2
    button = ui.NewButton("Select File 2")
    entry2 := ui.NewEntry()
    entry2.SetReadOnly(true)
    button.OnClicked(func(*ui.Button) {
        filename2 := ui.OpenFile(mainwin)
        if filename2 == "" {
            filename2 = ""
        }
        entry2.SetText(filename2)
        colnames := getColNames(filename2)
        updateColList(fileCols, colnames)
    })

    grid.Append(button,
        0, 1, 1, 1,
        false, ui.AlignFill, false, ui.AlignFill)
    grid.Append(entry2,
        1, 1, 1, 1,
        true, ui.AlignFill, false, ui.AlignFill)

    // column select
    form := ui.NewForm()
    form.SetPadded(true)
    vbox.Append(form, false)
    form.Append("Select Column to Compare by:", fileCols, true)
    fileCols.OnSelected(func(*ui.Combobox) {
        indexcol = fileCols.Selected() + 1
    })
    vbox.Append(ui.NewHorizontalSeparator(), false)

    // output grid
    grid = ui.NewGrid()
    grid.SetPadded(true)
    vbox.Append(grid, false)

    button = ui.NewButton("Select Output File Location")
    entry3 := ui.NewEntry()
    entry3.SetReadOnly(true)
    button.OnClicked(func(*ui.Button) {
        filename3 := ui.SaveFile(mainwin)
        if filename3 == "" {
            filename3 = ""
        }
        entry3.SetText(filename3)
    })

    grid.Append(button,
        0, 0, 1, 1,
        false, ui.AlignFill, false, ui.AlignFill)
    grid.Append(entry3,
        1, 0, 1, 1,
        true, ui.AlignFill, false, ui.AlignFill)

    button = ui.NewButton("Start Comparison")
    pb := ui.NewProgressBar()

    button.OnClicked(func(*ui.Button) {
        pb.SetValue(-1)
        outval := false
        if (entry1.Text() == "") || (entry2.Text() == "") || (entry3.Text() == "") {
            ui.MsgBoxError(mainwin, "File Select Error", "Please select input and output files first.")
            return
        }
        outval = doTheThing(entry1.Text(), entry2.Text(), entry3.Text(), indexcol)
        if outval == true {
            pb.SetValue(100)
            msg := "Complete! Wrote result to " + entry3.Text()
            ui.MsgBox(mainwin, "Complete!", msg)
        } else {
            pb.SetValue(0)
            ui.MsgBoxError(mainwin, "Error", "Make sure all files are selected first.")
        }
    })

    grid.Append(button,
        0, 1, 1, 1,
        false, ui.AlignFill, false, ui.AlignFill)

    // progress grid
    grid = ui.NewGrid()
    grid.SetPadded(true)
    vbox.Append(grid, true)

    grid.Append(pb,
        0, 0, 1, 1,
        true, ui.AlignFill, false, ui.AlignFill)

    return hbox
}

// build the iterface
func setupUI() {
    mainwin = ui.NewWindow("Excel File Comparison Tool", 640, 480, true)
    mainwin.OnClosing(func(*ui.Window) bool {
        ui.Quit()
        return true
    })
    ui.OnShouldQuit(func() bool {
        mainwin.Destroy()
        return true
    })

    tab := ui.NewTab()
    mainwin.SetChild(tab)
    mainwin.SetMargined(true)

    tab.Append("Controls", makeControlsPage())
    tab.SetMargined(0, true)

    mainwin.Show()

}

func main() {
    ui.Main(setupUI)
}
