# ged
Ged, short for git excel diffing tool, is a tool used for diffing two excel workbooks.

## Installation
Download the `gedInstaller_<version number>.exe` run the installer and then add
ged to the system path. The default install location is `C:\Program Files (x86)\ged`.

## Usage
### Basic Usage
```
ged <excelfilename>.xlsx
```
This will generate a html file called `<excelfilename>-diff.html` that will show
the differences between the local file and the one on the default branch. Open
this diff.html file in a web browser and view the differences between the two excel files.

### Bringing up the help menu
There are two ways to bring up the help menu typing `ged` by itself or `ged -h`

## How it works
The "smart compare" works by using a unique primary key in each excel sheet. This key
is either provided by the user or ged will try and find one by combining any
number of columns together. If they key repeats or ged is not able to find a
unique primary key then it will default to the basic diffing algorithm that finds
the rows that don't exist or changed between the "mine" and "theirs" sheets and
prints them out to the diff.
