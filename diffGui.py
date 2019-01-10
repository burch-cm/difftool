import easygui
import subprocess

file1 = easygui.fileopenbox(msg = "Select file 1 to compare", default = "\*.xlsx", filetypes = ["\*.xlsx"])
file2 = easygui.fileopenbox(msg = "Select file 2 to compare", default = "\*.xlsx", filetypes = ["\*.xlsx"])
outfile = easygui.filesavebox(msg = "Save results as...", default = "comparison_results.xlsx")

call = ["diffTool.exe",
		"-file1=" + file1,
		"-file2=" + file2,
		"-indexcol=6",
		"-outfile=" + outfile]

subprocess.run(call)
easygui.msgbox("Process completed!")