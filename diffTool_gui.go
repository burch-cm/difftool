package main

import (
	"github.com/andlabs/ui"
	_ "github.com/andlabs/ui/winmanifest"
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
	form.Append("Key Column (column to match):", fileCols, true)
	fileCols.OnSelected(func(*ui.Combobox) {
		indexcol = fileCols.Selected() + 1
	})
	vbox.Append(ui.NewHorizontalSeparator(), false)

	// output grid
	grid = ui.NewGrid()
	grid.SetPadded(true)
	vbox.Append(grid, false)
	// save location
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
		go pb.SetValue(-1)
		outval := false
		if (entry1.Text() == "") || (entry2.Text() == "") || (entry3.Text() == "") {
			ui.MsgBoxError(mainwin, "File Select Error", "Please select input and output files first.")
			return
		}
		fileDiff, _ := Difference(entry1.Text(), entry2.Text(), indexcol)
		outval = fileDiff.writeFile(entry3.Text())
		go pb.SetValue(100)
		if outval == true {

			msg := "Complete! Wrote result to " + entry3.Text()
			ui.MsgBox(mainwin, "Complete!", msg)
		} else {
			go pb.SetValue(0)
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
	mainwin = ui.NewWindow("diffTool v0.6.1", 640, 480, true)
	mainwin.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	ui.OnShouldQuit(func() bool {
		mainwin.Destroy()
		return true
	})

	tab1 := ui.NewTab()
	mainwin.SetChild(tab1)
	mainwin.SetMargined(true)

	tab1.Append("Controls", makeControlsPage())
	tab1.SetMargined(0, true)

	mainwin.Show()

}

func main() {
	ui.Main(setupUI)
}
