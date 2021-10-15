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
	SymLinkPath      string
	ModificationTime Date
	SubFolder        []FileData // If it's folder, here we save all children files data
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
	Total     int
	FileNames []string
	MainData  FileData
}

var STARTDIR string

func main() {
	STARTDIR, _ = os.Getwd()

	flagsToUse, paths, files, folders, wasPath := CollectElements(os.Args)

	var contentList []FileData
	var ts []FolderContent

	contentList = ReadDir(STARTDIR, contentList, flagsToUse.Flag_a, flagsToUse.Flag_R)
	contentList = ApplyFlags(flagsToUse, contentList)

	if (len(paths) == 0 && len(folders) == 0 && !wasPath) || len(files) != 0 {
		DataFromMainDir(files, contentList, flagsToUse, &ts)
	}

	DataFromDifferentDir(paths, flagsToUse, &ts)
	DataFromDifferentDir(folders, flagsToUse, &ts)

	PrintData(ts, flagsToUse)
}

func ReadDir(path string, content []FileData, skipHidden bool, RFlag bool) []FileData {
	var fileList []FileData

	os.Chdir(path)
	saveDirPath := path

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

	file, err := os.Open(".")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	SortWordArr(list)

	for _, name := range list {
		if IsHidden(name) && !skipHidden {
			continue
		}

		var dataToAppend FileData
		CollectData(&dataToAppend, name, saveDirPath)

		if dataToAppend.Name == "" {
			return content
		}

		if dataToAppend.isDirectory && RFlag {
			subFolderPath := path + "/" + dataToAppend.Name
			dataToAppend.SubFolder = ReadDir(subFolderPath, content, skipHidden, RFlag)
		}

		os.Chdir(saveDirPath)
		fileList = append(fileList, dataToAppend)
		file.Close()
	}

	return fileList
}

func CollectData(dataToAppend *FileData, name string, saveDirPath string) {
	fileInfo, err := os.Lstat(name)

	if err != nil {
		return
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		dataToAppend.SymLinkPath, _ = os.Readlink(name)
	}

	timeToAppend := fmt.Sprintf("%+03d:%+03d", fileInfo.ModTime().Hour(), fileInfo.ModTime().Minute())
	timeToAppend = strings.Replace(timeToAppend, "+", "", -1)

	dataToAppend.isDirectory = fileInfo.IsDir()
	dataToAppend.isHidden = IsHidden(name)
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

func IsHidden(filename string) bool {
	if filename[0:1] == "." {
		return true
	}
	return false
}

func CollectElements(arguments []string) (Flags, []string, []string, []string, bool) {
	var flagsToUse Flags
	var paths []string
	var files []string
	var folders []string

	var wasPath bool

	inputArgs := arguments[1:]

	for i := 0; i < len(inputArgs); i++ {
		if inputArgs[i][:1] == "/" {
			if CheckPath(inputArgs[i]) {
				paths = append(paths, inputArgs[i])
				continue
			} else {
				fmt.Printf("ls: cannot access '%s': No such file or directory\n", inputArgs[i])
				wasPath = true
				continue
			}
		}

		if strings.Contains(inputArgs[i], "/") {
			if CheckPath(STARTDIR + "/" + inputArgs[i]) {
				folders = append(folders, STARTDIR+"/"+inputArgs[i])
				continue
			} else {
				fmt.Printf("ls: cannot access '%s': Not a directory\n", inputArgs[i])
				wasPath = true
				continue
			}
		}

		if !strings.Contains(inputArgs[i], "-") || inputArgs[i] == "-" {
			fileInfo, err := os.Lstat(inputArgs[i])

			if err == nil {
				if fileInfo.Mode()&os.ModeSymlink == 0 {
					if CheckPath(STARTDIR + "/" + inputArgs[i]) {
						folders = append(folders, STARTDIR+"/"+inputArgs[i])
						continue
					}
				}
			}

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

	return flagsToUse, paths, files, folders, wasPath
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

func DataFromMainDir(files []string, contentList []FileData, flagsToUse Flags, ts *[]FolderContent) {
	if len(files) != 0 {
		var seekingContent []FileData
		SortWordArr(files)

		for i := 0; i < len(files); i++ {
			for k := 0; k < len(contentList); k++ {
				if contentList[k].Name == files[i] {
					seekingContent = append(seekingContent, contentList[k])
					break
				}

				if k == len(contentList)-1 {
					fmt.Printf("ls: cannot access '%s': No such file or directory\n", files[i])
				}
			}
		}

		seekingContent = ApplyFlags(flagsToUse, seekingContent)
		if len(seekingContent) != 0 {
			CollectFiles(seekingContent, seekingContent[0].Path, flagsToUse, ts, true)
		}
	} else {
		CollectFiles(contentList, contentList[0].Path, flagsToUse, ts, false)
	}
}

func DataFromDifferentDir(paths []string, flagsToUse Flags, ts *[]FolderContent) {
	for i := 0; i < len(paths); i++ {
		var tempVar []FileData
		tempVar = ReadDir(paths[i], tempVar, flagsToUse.Flag_a, flagsToUse.Flag_R)
		tempVar = ApplyFlags(flagsToUse, tempVar)
		CollectFiles(tempVar, tempVar[0].Path, flagsToUse, ts, false)
	}
}

func PrintData(ts []FolderContent, flagsToUse Flags) {
	for i := 0; i < len(ts); i++ {
		if len(ts) > 1 && !flagsToUse.Flag_R && ts[i].Path != "null" {
			fmt.Println(ts[i].Path)
		}

		for k := 0; k < len(ts[i].FileNames); k++ {
			if flagsToUse.Flag_R && k == 0 && ts[i].Path != "null" {
				fmt.Println(ts[i].Path)
			}
			if flagsToUse.Flag_l && k == 0 && ts[i].Total != -1 {
				fmt.Println("total:", ts[i].Total)
			}

			fmt.Print(ts[i].FileNames[k])
		}

		if len(ts) == 1 && !flagsToUse.Flag_l && len(ts[i].FileNames) != 0 {
			fmt.Println()
		}

		if len(ts) > 1 && !flagsToUse.Flag_l {
			fmt.Println()
		}

		if i != len(ts)-1 {
			fmt.Println()
		}
	}
}

func ApplyFlags(flagsToUse Flags, contentList []FileData) []FileData {
	if flagsToUse.Flag_t == true {
		SortByTime(contentList)
	}
	if flagsToUse.Flag_r == true {
		contentList = ReverseList(contentList)
	}
	return contentList
}

func CollectFiles(content []FileData, path string, flagsToUse Flags, res *[]FolderContent, certainFiles bool) {
	var dataToAppend FolderContent
	var totalCalculated bool

	if certainFiles {
		dataToAppend.Path = "null"
	} else {
		dataToAppend.Path = path + ":"
	}

	for i := 0; i < len(content); i++ {
		if !flagsToUse.Flag_a && content[i].isHidden {
			if content[i].Name == "." {
				dataToAppend.MainData = content[i]
			}
			continue
		}

		if flagsToUse.Flag_l {
			if !totalCalculated {
				if certainFiles {
					dataToAppend.Total = -1
				} else {
					dataToAppend.Total = CalculateBlocks(content)
				}
				totalCalculated = true
			}

			var fileName string = content[i].Permission + " " + strconv.Itoa(content[i].Hardlinks) + " " + content[i].Owner + " " + content[i].Group + " " + strconv.Itoa(int(content[i].Size)) + " " + content[i].ModificationTime.Month.UTC().Format("Jan") + " " + strconv.Itoa(content[i].ModificationTime.Day) + " " + content[i].ModificationTime.Time + " " + content[i].Name + "\n"

			if content[i].SymLinkPath != "" {
				fileName = fileName[:len(fileName)-1]
				fileName = fileName + " -> " + content[i].SymLinkPath + "\n"
			}

			dataToAppend.FileNames = append(dataToAppend.FileNames, fileName)

		} else {
			dataToAppend.FileNames = append(dataToAppend.FileNames, content[i].Name+" ")

			// if i == len(content)-1 {
			// 	dataToAppend.FileNames = append(dataToAppend.FileNames, "\n")
			// }
		}
	}

	*res = append(*res, dataToAppend)

	if flagsToUse.Flag_R {
		for i := 0; i < len(content); i++ {
			if content[i].isDirectory && content[i].Name != "." && content[i].Name != ".." {
				CollectFiles(content[i].SubFolder, path+"/"+content[i].Name, flagsToUse, res, false)
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
					} else if table[j].ModificationTime.Month.Hour() == table[j+1].ModificationTime.Month.Hour() {
						if table[j].ModificationTime.Month.Minute() < table[j+1].ModificationTime.Month.Minute() {
							tempVar := table[j]
							table[j] = table[j+1]
							table[j+1] = tempVar
						} else if table[j].ModificationTime.Month.Minute() == table[j+1].ModificationTime.Month.Minute() {
							if table[j].ModificationTime.Month.Second() < table[j+1].ModificationTime.Month.Second() {
								tempVar := table[j]
								table[j] = table[j+1]
								table[j+1] = tempVar
							} else if table[j].ModificationTime.Month.Second() == table[j+1].ModificationTime.Month.Second() {
								if table[j].ModificationTime.Month.Nanosecond() < table[j+1].ModificationTime.Month.Nanosecond() {
									tempVar := table[j]
									table[j] = table[j+1]
									table[j+1] = tempVar
								}
							}
						}
					}
				}
			}
		}
	}

	for i := 0; i < len(table); i++ {
		if table[i].isDirectory {
			SortByTime(table[i].SubFolder)
		}
	}
}

func ReverseList(table []FileData) []FileData {
	var result []FileData

	for i := len(table) - 1; 0 <= i; i-- {
		result = append(result, table[i])
	}

	for i := 0; i < len(table); i++ {
		if result[i].isDirectory {
			result[i].SubFolder = ReverseList(result[i].SubFolder)
		}
	}

	return result
}

func GetUpperPath(path string) string {
	pathInRune := []rune(path)
	var dashCounter int = 0

	for k := 0; k < len(pathInRune); k++ {
		if pathInRune[k] == '/' {
			dashCounter++
		}
	}

	if dashCounter == 1 {
		return "/"
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

func CheckPath(path string) bool {
	err := os.Chdir(path)
	if err != nil {
		os.Chdir(STARTDIR)
		return false
	}

	return true
}
