package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/mxschmitt/golang-combinations"
	"github.com/xuri/excelize/v2"
)

type differenceType int

const (
	difference differenceType = iota
	addition                  = iota
)

type differenceLine struct {
	key      string
	lineType differenceType
	linePos  int
}

// write Sheets to csv
func writeSheetsToCsv(excelFile *excelize.File, mine bool) []string {
	excelSheetNames := excelFile.GetSheetList()

	var sheetNames []string

	for _, sheet := range excelSheetNames {
		rows, err := excelFile.GetRows(sheet)
		if err != nil {
			panic(err)
		}

		maxRowLen := 0
		for _, row := range rows {

			if len(row) > maxRowLen {
				maxRowLen = len(row)
			}
		}

		sheetNames = append(sheetNames, sheet)

		csvFileName := getSheetFileName(sheet, mine)
		csvFile, err := os.Create(csvFileName)

		if err != nil {
			panic(err)
		}

		csvwriter := csv.NewWriter(csvFile)

		for _, row := range rows {
			if len(row) < maxRowLen {
				for len(row) < maxRowLen {
					row = append(row, "")
				}
			}
			csvwriter.Write(row)
			csvwriter.Flush()
		}

		csvFile.Close()
	}

	return sheetNames

}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func removeFiles(sheets []string, mine bool) {
	for _, sheet := range sheets {
		err := os.Remove(getSheetFileName(sheet, mine))
		if err != nil {
			fmt.Print(err)
		}
	}
}

func concatSheetNames(theirsSheets, mineSheets []string) []string {
	var sheets []string

	for _, sheet := range theirsSheets {
		sheets = append(sheets, sheet)
	}

	for _, sheet := range mineSheets {
		if !containsString(sheets, sheet) {
			sheets = append(sheets, sheet)
		}
	}

	return sheets
}

func getSheetFileName(sheet string, mine bool) string {
	if mine {
		return "mine-" + sheet + "-ABdiffTool.csv"
	}
	return "theirs-" + sheet + "-ABdiffTool.csv"
}

func sanatizeKeys(rawKeys []string) []string {
	keys := []string{}

	for _, key := range rawKeys {
		if key != "" {
			keys = append(keys, key)
		}
	}

	return keys
}

func autoFindPrimaryKeyNames(dataMine [][]string, dataTheirs [][]string) []string {
	// checks to make sure there is data to compare
	if len(dataTheirs) == 0 || len(dataMine) == 0 {
		return []string{}
	}

	if verboseOutput {
		fmt.Printf("Selecting from raw keys (%d): %s\n", len(dataMine[0]), dataMine[0])
	}

	possibleKeys := sanatizeKeys(dataMine[0])

	if verboseOutput {
		fmt.Printf("Sanitized Keys (%d): %s\n", len(possibleKeys), possibleKeys)
	}

	for length := 1; length <= len(possibleKeys); length++ {
		for _, keys := range combinations.Combinations(possibleKeys, length) {
			if verboseOutput {
				fmt.Printf("Trying key Combo: %s\n", keys)
			}
			indexesMine := findPrimaryKeyIndexes(dataMine[0], keys)
			if primaryKeysUnique(dataMine, indexesMine) == nil {
				indexesTheirs := findPrimaryKeyIndexes(dataTheirs[0], keys)
				if !reflect.DeepEqual(indexesMine, indexesTheirs) {
					if verboseOutput {
						fmt.Printf("Indexes not equal\n")
					}
					continue
				}
				if primaryKeysUnique(dataTheirs, indexesTheirs) == nil {
					return keys
				}
			}
		}

	}
	return []string{}
}

func findPrimaryKeyIndexes(header []string, primaryKeys []string) []int {
	keys := []int{}
	for index, title := range header {
		if slices.Contains(primaryKeys, title) {
			keys = append(keys, index)
		}
	}

	return keys
}

func concatKeysData(row []string, keyIndexes []int) string {
	var key string

	for index := range keyIndexes {
		key += "%@!#!@%" + row[keyIndexes[index]]
	}

	return key
}

func primaryKeysUnique(data [][]string, keyIndexes []int) error {
	for index, rowData := range data {
		key := concatKeysData(rowData, keyIndexes)
		for _, rowData := range data[index+1:] {
			if key == concatKeysData(rowData, keyIndexes) {
				if verboseOutput {
					fmt.Printf("Keys: %s found multiple times\n", key)
				}
				errorString := "key " + key + " found multiple times"
				return errors.New(errorString)
			}
		}
	}
	return nil
}

func createDataMap(data [][]string, keyIndexes []int) map[string][]string {
	index := 0
	maxLen := len(data)

	dataMap := make(map[string][]string)

	for index < maxLen {

		dataMap[concatKeysData(data[index], keyIndexes)] = data[index]

		index++
	}

	return dataMap
}

func normalizeData(dataTheirs, dataMine [][]string) ([][]string, [][]string) {
	if len(dataTheirs) == 0 || len(dataMine) == 0 {
		return dataTheirs, dataMine
	}

	dataLenTheirs := len(dataTheirs[0])
	dataLenMine := len(dataMine[0])

	if dataLenTheirs == dataLenMine {
		return dataTheirs, dataMine
	}

	if dataLenTheirs < dataLenMine {
		difference := dataLenMine - dataLenTheirs

		for index := range dataTheirs {
			for i := 0; i < difference; i++ {
				dataTheirs[index] = append(dataTheirs[index], "")
			}
		}
	}

	if dataLenMine < dataLenTheirs {
		difference := dataLenTheirs - dataLenMine

		for index := range dataMine {
			for i := 0; i < difference; i++ {
				dataMine[index] = append(dataMine[index], "")
			}
		}
	}

	return dataTheirs, dataMine
}

func listKeys(dataMap map[string][]string) []string {
	var keyList []string
	for key := range dataMap {
		keyList = append(keyList, key)
	}

	return keyList
}

func findMissing(baseList []string, compareList []string) []string {
	var missingKeys []string

	for _, key := range baseList {
		if !containsString(compareList, key) {
			missingKeys = append(missingKeys, key)
		}
	}

	return missingKeys
}

func rowsToStrings(data [][]string) []string {
	var rowStrings []string

	for _, row := range data {
		rowStrings = append(rowStrings, strings.Join(row, " "))
	}

	return rowStrings
}

func findDiffRows(dataTheirs [][]string, dataMine [][]string) ([][]string, [][]string) {
	var mineDiff [][]string
	var theirsDiff [][]string

	theirRowStrings := rowsToStrings(dataTheirs)
	mineRowStrings := rowsToStrings(dataMine)

	for _, row := range dataTheirs {
		if !containsString(mineRowStrings, strings.Join(row, " ")) {
			theirsDiff = append(theirsDiff, row)
		}
	}

	for _, row := range dataMine {
		if !containsString(theirRowStrings, strings.Join(row, " ")) {
			mineDiff = append(mineDiff, row)
		}
	}

	return theirsDiff, mineDiff
}

func findDifferences(theirsDataMap, mineDataMap map[string][]string) []string {
	var differentRows []string

	for key := range theirsDataMap {
		if strings.Join(theirsDataMap[key], " ") != strings.Join(mineDataMap[key], " ") {
			differentRows = append(differentRows, key)
		}
	}

	return differentRows
}

// TODO: add more intelligence for additional lines to be in correct order
func orderAndTypeDiffLines(missingFromTheirs []string, differentKeys []string, dataTheirs [][]string, dataMine [][]string, primaryKeyIndexes []int) []differenceLine {
	var lineDiffs []differenceLine

	for _, key := range missingFromTheirs {
		for index, line := range dataMine {
			if key == concatKeysData(line, primaryKeyIndexes) {
				lineDiffs = append(lineDiffs, differenceLine{key: key, lineType: addition, linePos: index})
				break
			}
		}
	}

	for _, key := range differentKeys {
		for index, line := range dataTheirs {
			if key == concatKeysData(line, primaryKeyIndexes) {
				lineDiffs = append(lineDiffs, differenceLine{key: key, lineType: difference, linePos: index})
				break
			}
		}
	}

	sort.SliceStable(lineDiffs, func(i, j int) bool {
		return lineDiffs[i].linePos < lineDiffs[j].linePos
	})

	return lineDiffs
}

func compareCSV(dataTheirs [][]string, dataMine [][]string, primaryKeys []string, sheetName string, htmlFile *os.File, smartCompare bool) {

	if !smartCompare {
		fmt.Printf("Smart compare turned off using default diff algorithm for %s\r\n", sheetName)
	}

	if len(primaryKeys) == 0 && smartCompare {
		fmt.Printf("Attempting to find primary key for %s\r\n", sheetName)
		primaryKeys = autoFindPrimaryKeyNames(dataMine, dataTheirs)
		if len(primaryKeys) == 0 {
			fmt.Printf("Unable to find suitable primary key. Using default diff algorithm for %s\r\n", sheetName)
			smartCompare = false
		} else {
			fmt.Printf("Primary key %s found for %s\r\n", primaryKeys, sheetName)
		}
	}

	var theirsPrimaryKeyIndexes []int
	if len(dataTheirs) > 0 && smartCompare {
		theirsPrimaryKeyIndexes = findPrimaryKeyIndexes(dataTheirs[0], primaryKeys)
	}

	var minePrimaryKeyIndexes []int
	if len(dataMine) > 0 && smartCompare {
		minePrimaryKeyIndexes = findPrimaryKeyIndexes(dataMine[0], primaryKeys)
	}

	if !reflect.DeepEqual(minePrimaryKeyIndexes, theirsPrimaryKeyIndexes) {
		fmt.Printf("Primary key indexes don't match. Using default diff algorithm for %s\r\n", sheetName)
		smartCompare = false
	}

	if (len(theirsPrimaryKeyIndexes) == 0 || len(minePrimaryKeyIndexes) == 0) && smartCompare {
		fmt.Printf("Unable to find primary key. Using default diff algorithm for %s\r\n", sheetName)
		smartCompare = false
	}

	dataTheirs, dataMine = normalizeData(dataTheirs, dataMine)

	// check if keys are primary only if the previous checks passed
	if smartCompare {
		err := primaryKeysUnique(dataTheirs, theirsPrimaryKeyIndexes)
		if err != nil {
			fmt.Printf("%s theirs: %s\n", sheetName, err)
			smartCompare = false
		}

		err = primaryKeysUnique(dataMine, minePrimaryKeyIndexes)

		if err != nil {
			fmt.Printf("%s mine: %s\n", sheetName, err)
			smartCompare = false
		}
	}

	if smartCompare {
		theirsDataMap := createDataMap(dataTheirs, theirsPrimaryKeyIndexes)
		mineDataMap := createDataMap(dataMine, minePrimaryKeyIndexes)

		if reflect.DeepEqual(theirsDataMap, mineDataMap) {
			htmlFile.Write([]byte(htmlAddSheetHeader(sheetName)))
			htmlFile.Write([]byte(htmlAddSubHeading("Sheets are equal")))
			htmlFile.Write([]byte(htmlAddBreakLine()))
			return
		}

		mineKeylist := listKeys(mineDataMap)
		theirsKeylist := listKeys(theirsDataMap)

		missingFromTheirs := findMissing(mineKeylist, theirsKeylist)

		differentKeys := findDifferences(theirsDataMap, mineDataMap)

		lineDifferences := orderAndTypeDiffLines(missingFromTheirs, differentKeys, dataTheirs, dataMine, minePrimaryKeyIndexes)

		htmlFile.Write([]byte(htmlAddSheetHeader(sheetName)))
		htmlFile.Write([]byte(htmlStartTable()))

		htmlFile.Write([]byte(htmlAddTableHeaderDefaultDiff(dataMine[0])))

		for _, diffLine := range lineDifferences {
			htmlFile.Write([]byte(htmlAddDiffRow(theirsDataMap, mineDataMap, diffLine)))
		}

		htmlFile.Write([]byte(htmlEndTable()))
		htmlFile.Write([]byte(htmlAddBreakLine()))
	} else {
		theirDiff, mineDiff := findDiffRows(dataTheirs, dataMine)

		if theirDiff == nil && mineDiff == nil {
			htmlFile.Write([]byte(htmlAddSheetHeader(sheetName)))
			htmlFile.Write([]byte(htmlAddSubHeading("Sheets are equal")))
			htmlFile.Write([]byte(htmlAddBreakLine()))
			return
		}

		htmlFile.Write([]byte(htmlAddSheetHeader(sheetName)))
		htmlFile.Write([]byte(htmlAddSubHeading("Mine")))
		if len(mineDiff) > 0 {
			htmlFile.Write([]byte(htmlStartTable()))
			htmlFile.Write([]byte(htmlAddTableHeaderDefaultDiff(dataMine[0])))

			for _, diffLine := range mineDiff {
				htmlFile.Write([]byte(htmlAddRow(diffLine)))
			}

			htmlFile.Write([]byte(htmlEndTable()))
		}

		htmlFile.Write([]byte(htmlAddSubHeading("Theirs")))
		if len(theirDiff) > 0 {
			htmlFile.Write([]byte(htmlStartTable()))
			htmlFile.Write([]byte(htmlAddTableHeaderDefaultDiff(dataTheirs[0])))

			for _, diffLine := range theirDiff {
				htmlFile.Write([]byte(htmlAddRow(diffLine)))
			}

			htmlFile.Write([]byte(htmlEndTable()))
		}
		htmlFile.Write([]byte(htmlAddBreakLine()))

	}
}
