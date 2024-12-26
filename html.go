package main

import ()

func htmlAddSheetHeader(sheetName string) string {
	return "<header align=\"center\"><h1>" + sheetName + "</h1></header>"
}

func htmlAddSubHeading(value string) string {
	return "<header align=\"center\"><h2>" + value + "</h2></header>"
}

func htmlStartTable() string {
	return "<table align=\"center\" border=\"1\">"
}

func htmlEndTable() string {
	return "</table>\n"
}

func htmlAddBreakLine() string {
	return "<hr>\n"
}

func htmlAddDiffRow(theirsDataMap, mineDataMap map[string][]string, diffLine differenceLine) string {
	rowString := "<tr>"

	// handle if line is an addition
	if diffLine.lineType == addition {
		return htmlAddNewRow(mineDataMap, diffLine.key)
	}

	theirRow := theirsDataMap[diffLine.key]
	mineRow := mineDataMap[diffLine.key]

	for index, cell := range theirRow {
		if mineRow == nil {
			rowString += "<td align=\"center\" style=\"padding:10px\"><font color=\"red\">" + cell + "</font></td>"
		} else if cell != mineRow[index] {
			rowString += "<td align=\"center\" style=\"padding:10px\"><font color=\"red\">" + cell + "</font><br /><font color=\"green\">" + mineRow[index] + "</font></td>"
		} else {
			rowString += "<td align=\"center\" style=\"padding:10px\">" + cell + "</td>"
		}
	}

	rowString += "</tr>\n"

	return rowString
}

func htmlAddNewRow(mineDataMap map[string][]string, key string) string {
	rowString := "<tr>"

	mineRow := mineDataMap[key]

	for _, cell := range mineRow {
		rowString += "<td align=\"center\" style=\"padding:10px\"><font color=\"green\">" + cell + "</font></td>"
	}

	rowString += "</tr>\n"

	return rowString
}

func htmlAddRow(rowData []string) string {
	rowString := "<tr>"

	for _, cell := range rowData {
		rowString += "<td align=\"center\" style=\"padding:10px\">" + cell + "</td>"
	}

	rowString += "</tr>\n"

	return rowString
}

func htmlAddTableHeader(dataMap map[string][]string, primaryKey string) string {
	rowString := "<tr>"

	row := dataMap[primaryKey]
	for _, cell := range row {
		rowString += "<td align=\"center\" style=\"padding:10px\"><b>" + cell + "</b></td>"
	}

	rowString += "</tr>\n"

	return rowString
}

func htmlAddTableHeaderDefaultDiff(rowData []string) string {
	rowString := "<tr>"

	for _, cell := range rowData {
		rowString += "<td align=\"center\" style=\"padding:10px\"><b>" + cell + "</b></td>"
	}

	rowString += "</tr>\n"

	return rowString
}
