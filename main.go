package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type FileData struct {
	isDirectory      bool
	isHidden         bool
	Permission       string
	Name             string
	Hardlinks        int
	Owner            string
	Group            string
	Size             int64
	SizeKB           int
	ModificationTime Date
	SubFolder        []FileData
	Path             string
}

type Date struct {
	Month time.Time
	Day   int
	Time  string
}

type Flags struct {
	Flag_l bool
	Flag_R bool
	Flag_a bool
	Flag_r bool
	Flag_t bool
}

type FolderContent struct {
	Path      string
	FilesData []FileData
}

var STARTDIR string
var USERHOMEDIR string

func ReadDir(path string, content []FolderContent, skipHidden bool) []FolderContent {
	var fileList []FileData

	os.Chdir(path)
	saveDirPath := path

	if skipHidden {
		var pathToWorkWith string
		var name string
		UpperPath := GetUpperPath(path)

		for m := 0; m < 2; m++ {
			var DotsFolderData FileData
			if m == 0 {
				pathToWorkWith = path
				name = "."
			} else {
				pathToWorkWith = UpperPath
				name = ".."
			}

			CollectData(&DotsFolderData, pathToWorkWith, saveDirPath)
			DotsFolderData.Name = name
			DotsFolderData.isHidden = true
			fileList = append(fileList, DotsFolderData)
		}
	}

	file, err := os.Open(".")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	SortWordArr(list)

	for _, name := range list {
		if isHidden(name) && !skipHidden {
			continue
		}

		var dataToAppend FileData
		CollectData(&dataToAppend, name, saveDirPath)

		if dataToAppend.Name == "" {
			return content
		}

		if dataToAppend.isDirectory {
			subFolderPath := path + "/" + dataToAppend.Name
			content = ReadDir(subFolderPath, content, skipHidden)
		}

		os.Chdir(saveDirPath)
		fileList = append(fileList, dataToAppend)
	}

	if len(fileList) != 0 {
		var contentToAppend FolderContent
		contentToAppend.Path = fileList[0].Path
		contentToAppend.FilesData = fileList
		content = append(content, contentToAppend)
	}

	return content
}

func CollectData(dataToAppend *FileData, name string, saveDirPath string) {
	fileInfo, err := os.Stat(name)

	if err != nil {
		return
	}

	timeToAppend := fmt.Sprintf("%+03d:%+03d", fileInfo.ModTime().Hour(), fileInfo.ModTime().Minute())
	timeToAppend = strings.Replace(timeToAppend, "+", "", -1)

	dataToAppend.isDirectory = fileInfo.IsDir()
	dataToAppend.isHidden = isHidden(name)
	dataToAppend.Permission = fmt.Sprintf("%v", fileInfo.Mode())
	dataToAppend.Name = fileInfo.Name()
	dataToAppend.Size = fileInfo.Size()
	dataToAppend.ModificationTime.Day = fileInfo.ModTime().Day()
	dataToAppend.ModificationTime.Month = fileInfo.ModTime()
	dataToAppend.ModificationTime.Time = timeToAppend
	dataToAppend.Path = saveDirPath

	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		UID, _ := user.LookupId(strconv.Itoa(int(stat.Uid)))
		GID, _ := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

		dataToAppend.Hardlinks = int(stat.Nlink)
		dataToAppend.SizeKB = int(stat.Blocks / 2)
		dataToAppend.Owner = UID.Username
		dataToAppend.Group = GID.Name
	}
}

func isHidden(filename string) bool {
	if filename[0:1] == "." {
		return true
	}

	return false
}

func CollectElements(arguments []string) (Flags, []string, []string) {
	var flagsToUse Flags
	var paths []string
	var files []string

	inputArgs := arguments[1:]

	for i := 0; i < len(inputArgs); i++ {
		if strings.Contains(inputArgs[i], "/") {
			paths = append(paths, inputArgs[i])
			continue
		}

		if inputArgs[i] == "-" {
			paths = append(paths, inputArgs[i])
			continue
		}

		if !strings.Contains(inputArgs[i], "-") {
			files = append(files, inputArgs[i])
			continue
		}

		for _, k := range inputArgs[i] {
			if k == '-' {
				flagsToUse = DetectFlag(inputArgs[i], flagsToUse)
				break
			} else {
				paths = append(paths, inputArgs[i])
				break
			}
		}
	}

	return flagsToUse, paths, files
}

func DetectFlag(flagToCheck string, flagsToUse Flags) Flags {
	dashFound := false
	for _, i := range flagToCheck {
		if i == 'R' && flagsToUse.Flag_R == false {
			flagsToUse.Flag_R = true
		} else if i == 'a' && flagsToUse.Flag_a == false {
			flagsToUse.Flag_a = true
		} else if i == 'r' && flagsToUse.Flag_r == false {
			flagsToUse.Flag_r = true
		} else if i == 'l' && flagsToUse.Flag_l == false {
			flagsToUse.Flag_l = true
		} else if i == 't' && flagsToUse.Flag_t == false {
			flagsToUse.Flag_t = true
		} else if i == '-' && !dashFound {
			dashFound = true
		} else {
			fmt.Println("ERROR")
			os.Exit(0)
		}
	}

	return flagsToUse
}

func main() {
	var contentList []FolderContent

	STARTDIR, _ = os.Getwd()
	flagsToUse, paths, files := CollectElements(os.Args)
	contentList = ReadDir(STARTDIR, contentList, flagsToUse.Flag_a)

	_ = paths
	_ = files

	SortByPath(contentList)
	if flagsToUse.Flag_t == true {
		for i := 0; i < len(contentList); i++ {
			SortByTime(contentList[i].FilesData)
		}
	}

	if flagsToUse.Flag_r == true {

	}

	if flagsToUse.Flag_R == true {
		PrintFolders(contentList, flagsToUse)
	} else {
		PrintFiles(contentList[0].FilesData, flagsToUse)
	}
}

func PrintFolders(list []FolderContent, flagsToUse Flags) {
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i].Path + ":")
		PrintFiles(list[i].FilesData, flagsToUse)
		if i != len(list)-1 {
			fmt.Println()
		}
	}
}

func PrintExactFolder() {

}

func PrintFiles(fileNames []FileData, flagsToUse Flags) {
	for i := 0; i < len(fileNames); i++ {
		if flagsToUse.Flag_l {
			if i == 0 {
				fmt.Println("total", CalculateBlocks(fileNames))
			}
			fmt.Println(fileNames[i].Permission, fileNames[i].Hardlinks, fileNames[i].Owner, fileNames[i].Group, fileNames[i].Size, fileNames[i].ModificationTime.Month.UTC().Format("Jan"), fileNames[i].ModificationTime.Day, fileNames[i].ModificationTime.Time, fileNames[i].Name)
		} else {
			fmt.Print(fileNames[i].Name + " ")
			if i == len(fileNames)-1 {
				fmt.Println()
			}
		}
	}
}

func CalculateBlocks(filesData []FileData) int {
	var sum int
	for i, _ := range filesData {
		sum = sum + filesData[i].SizeKB
	}
	return sum
}

func SortByPath(table []FolderContent) {
	for i := 0; i < len(table); i++ {
		for j := 0; j < len(table)-i-1; j++ {
			if table[j].Path > table[j+1].Path {
				tempVar := table[j]
				table[j] = table[j+1]
				table[j+1] = tempVar
			}
		}
	}
}

func SortWordArr(table []string) {
	for i := 0; i < len(table); i++ {
		for j := 0; j < len(table)-i-1; j++ {
			if table[j] > table[j+1] {
				tempVar := table[j]
				table[j] = table[j+1]
				table[j+1] = tempVar
			}
		}
	}
}

func SortByTime(table []FileData) {
	for i := 0; i < len(table); i++ {
		for j := 0; j < len(table)-i-1; j++ {
			if table[j].ModificationTime.Month.Month().String() < table[j+1].ModificationTime.Month.Month().String() {
				tempVar := table[j]
				table[j] = table[j+1]
				table[j+1] = tempVar
			} else if table[j].ModificationTime.Month.Month().String() == table[j+1].ModificationTime.Month.Month().String() {
				if table[j].ModificationTime.Day < table[j+1].ModificationTime.Day {
					tempVar := table[j]
					table[j] = table[j+1]
					table[j+1] = tempVar
				} else if table[j].ModificationTime.Day == table[j+1].ModificationTime.Day {
					if table[j].ModificationTime.Month.Hour() < table[j+1].ModificationTime.Month.Hour() {
						tempVar := table[j]
						table[j] = table[j+1]
						table[j+1] = tempVar
					} else if table[j].ModificationTime.Month.Minute() < table[j+1].ModificationTime.Month.Minute() {
						tempVar := table[j]
						table[j] = table[j+1]
						table[j+1] = tempVar
					}
				}
			}
		}
	}
}

func GetUpperPath(path string) string {
	if len(path) == 1 {
		return path
	}

	for k := 0; k < len(path); k++ {
		if path[len(path)-1:] == "/" {
			path = path[:len(path)-1]
			break
		} else {
			path = path[:len(path)-1]
		}
	}

	return path
}
