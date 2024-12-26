package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

const VERSION = "0.2.4"


var verboseOutput bool
var sheetsMine []string
var sheetsTheirs []string

func usageMessage() {
	fmt.Printf("Usage: ged [arguments] <excel workbook>\n")
	flag.PrintDefaults()
}

// allow user to define .xlsx sheet, theirs version and primary key
func main() {
    userConfig := getConfig()
	var commitFlag = flag.String("c", userConfig.DefaultCommit, "Commit to compare against")
	var keyFlag = flag.String("k", "", "The primary key to be used for diffing the excel sheet")
	var otherFileNameFlag = flag.String("r", "", "The path to the remote file starting from the root of the git repo. This should only be used if a comparing to a file of a different name on the remote commit.")
	var outputFlag = flag.String("o", "", "Path to directory where the output file should go. Default is the current working directory")
	var localCompareFlag = flag.String("lc", "", "Relative Path to local file to compare against")
	var smartCompareOffFlag = flag.Bool("sco", false, "Tells ged to turn of smart compare and ignore primary keys")
	var verboseFlag = flag.Bool("v", false, "Display verbose output")
	var aboutFlag = flag.Bool("about", false, "Display about page for ged")
    var newDefaultCommit = flag.String("setDefaultCommit", "", "Sets the default commit")

	flag.Usage = usageMessage
	flag.Parse()

	if *aboutFlag {
		fmt.Printf("ged version: %s\n\n", VERSION)
		fmt.Printf("Author: Austin Bunker, AB Engineering & Fabrication\n\n")
		os.Exit(0)
	}

	verboseOutput = *verboseFlag

    if *newDefaultCommit != "" {
        userConfig.DefaultCommit = *newDefaultCommit
        updateConfigFile(userConfig)
        os.Exit(0)
    }

	if len(flag.Args()) != 1 {
		fmt.Printf("Error: Incorrect number of positional arguments.\n")
		usageMessage()
		os.Exit(0)
	}

	if verboseOutput {
		fmt.Printf("CommitFlag: %s\r\n", *commitFlag)
		fmt.Printf("Key: %s\r\n", *keyFlag)
		fmt.Printf("otherFileNameFlag: %s\r\n", *otherFileNameFlag)
		fmt.Printf("outputFlag: %s\r\n", *outputFlag)
		fmt.Printf("localCompareFlag: %s\r\n", *localCompareFlag)
		fmt.Printf("SmartCompareOffFlag: %t\r\n", *smartCompareOffFlag)
		fmt.Printf("VerboseFlag: %t\r\n", *verboseFlag)
	}

	currentDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("Unable to find current directory. Aborting..\r\n")
		panic(err)
	}

	//get git root path if needed

	var gitRoot = []byte{}
	var gitRootString = ""

	if *localCompareFlag == "" {
		var err error = nil
		gitRoot, err = exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Printf("Unable to find git root folder. Aborting..\r\n")
			panic(err)
		}

		gitRootString = filepath.FromSlash(strings.TrimSpace(string(gitRoot)))
		gitRootString = gitRootString + string(filepath.Separator)
	}

	workBookGivenPath := flag.Arg(0)
	workBookFullPath := filepath.Join(currentDir, workBookGivenPath)
	mineWorkBookName := strings.ReplaceAll(filepath.Base(workBookFullPath), ".xlsx", "")

	var workBookGitPath = ""
	var theirWorkBookName = ""
	var theirWorkBook = ""

	//use the same name as mine-workbook
	if *otherFileNameFlag == "" && *localCompareFlag == "" {
		workBookGitPath = strings.ReplaceAll(workBookFullPath, gitRootString, "")
	} else if *localCompareFlag == "" { // use provided path
		workBookGitPath = *otherFileNameFlag
	} else {
		theirWorkBook = filepath.Join(currentDir, filepath.FromSlash(*localCompareFlag))
	}

	if *localCompareFlag == "" {
		theirWorkBookName = strings.ReplaceAll(filepath.Base(workBookGitPath), ".xlsx", "")

		theirWorkBook = filepath.Join(currentDir, theirWorkBookName+"-Theirs.xlsx")
	} else {
		theirWorkBookName = strings.ReplaceAll(filepath.Base(theirWorkBook), ".xlsx", "")
	}

	var outputFilePath = ""

	if *outputFlag == "" {
		outputFilePath = filepath.Join(currentDir, mineWorkBookName+"-diff.html")
	} else {
		outputDir := filepath.FromSlash(*outputFlag)
		outputFilePath = filepath.Join(currentDir, outputDir, mineWorkBookName+"-diff.html")
	}

	if verboseOutput {
		fmt.Printf("gitRoot: %s\n", gitRootString)
		fmt.Printf("workBookGitPath: %s\n", workBookGitPath)
		fmt.Printf("MineWorkBookName: %s\n", mineWorkBookName)
		fmt.Printf("TheirWorkBookName: %s\n", theirWorkBookName)
		fmt.Printf("TheirWorkBook: %s\n", theirWorkBook)
		fmt.Printf("workBookFullPath: %s\n", workBookFullPath)
		fmt.Printf("outputFilePath: %s\n", outputFilePath)
	}

	commit := *commitFlag

	if *localCompareFlag == "" {
		fmt.Printf("Diffing %s against their %s at %s\r\n", mineWorkBookName, theirWorkBookName, commit)
	} else {
		fmt.Printf("Diffing %s against their local %s\r\n", mineWorkBookName, theirWorkBookName)
	}

	primaryKey := *keyFlag
    primaryKeyList := []string{}
	if primaryKey != "" {
		fmt.Printf("Using primary key: %s\n", primaryKey)
        primaryKeyList = append(primaryKeyList, primaryKey)
    }

    // check to make sure there is something to compare against
    if *localCompareFlag == ""  && commit == ""{
        fmt.Printf("Error: No commit or file to compare against\n")
        os.Exit(0)
    }

	if *localCompareFlag == "" {
		theirWorkbookFile, err := os.Create(theirWorkBook)
		if err != nil {
			panic(err)
		}

		workBookGitPath = strings.ReplaceAll(workBookGitPath, "\\", "/")

		//get workbook theirs
		commandString := commit + ":" + workBookGitPath

		var stderr bytes.Buffer

		cmd := exec.Command("git", "show", commandString)
		cmd.Dir = gitRootString
		cmd.Stdout = theirWorkbookFile
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(stderr.String())
			theirWorkbookFile.Close()

			os.Remove(theirWorkBook)
			panic(err)
		}

		theirWorkbookFile.Close()
	}

	excelMine, err := excelize.OpenFile(workBookFullPath)
	if err != nil {
		panic(err)
	}

	defer func() {
		//close the spreadsheet
		if err := excelMine.Close(); err != nil {
			panic(err)
		}
	}()

	excelTheirs, err := excelize.OpenFile(theirWorkBook)

	if err != nil {
		panic(err)
	}

	defer func() {
		//close the spreadsheet
		if err := excelTheirs.Close(); err != nil {
			panic(err)
		}
	}()

	// write 'mine' sheets to a csv
	sheetsMine = writeSheetsToCsv(excelMine, true)

	// write 'theirs' sheets to a csv
	sheetsTheirs = writeSheetsToCsv(excelTheirs, false)

	htmlFile, err := os.Create(outputFilePath)
	if err != nil {
		removeFiles(sheetsMine, true)
		removeFiles(sheetsTheirs, false)
		panic(err)
	}

	combinedSheets := concatSheetNames(sheetsTheirs, sheetsMine)

	for _, sheetName := range combinedSheets {
		mineFile, err := os.Open(getSheetFileName(sheetName, true))
		if err != nil {
			mineFile, err = os.Create(getSheetFileName(sheetName, true))
			if err != nil {
				panic(err)
			}
		}
		csvReaderMine := csv.NewReader(mineFile)
		dataMine, err := csvReaderMine.ReadAll()
		if err != nil {
			panic(err)
		}
		mineFile.Close()

		theirsFile, err := os.Open(getSheetFileName(sheetName, false))
		if err != nil {
			theirsFile, err = os.Create(getSheetFileName(sheetName, false))
			if err != nil {
				panic(err)
			}
		}

		csvReaderTheirs := csv.NewReader(theirsFile)
		dataTheirs, err := csvReaderTheirs.ReadAll()
		if err != nil {
			panic(err)
		}
		theirsFile.Close()

		compareCSV(dataTheirs, dataMine, primaryKeyList, sheetName, htmlFile, !*smartCompareOffFlag)
	}

	htmlFile.Close()

	// remove temp csv files
	removeFiles(combinedSheets, true)
	removeFiles(combinedSheets, false)
	if *localCompareFlag == "" {
		err = os.Remove(theirWorkBook)
		if err != nil {
			panic(err)
		}
	}
}
