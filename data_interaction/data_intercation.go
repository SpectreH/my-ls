package data_interaction

import (
	"fmt"
	"log"
	"my-ls/calculations"
	"my-ls/checks"
	"my-ls/flags"
	"my-ls/sorts"
	"my-ls/structures"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

// Reads a certain directory to collect file names
func ReadDir(path string, content []structures.FileData, skipHidden bool, RFlag bool) []structures.FileData {
	var fileList []structures.FileData

	os.Chdir(path)
	saveDirPath := path

	var pathToWorkWith string
	var name string
	UpperPath := GetUpperPath(path)

	for m := 0; m < 2; m++ {
		var DotsFolderData structures.FileData
		if m == 0 {
			pathToWorkWith = path
			name = "."
		} else {
			pathToWorkWith = UpperPath
			name = ".."
		}

		AppendData(&DotsFolderData, pathToWorkWith, saveDirPath)
		DotsFolderData.Name = name
		DotsFolderData.IsHidden = true
		fileList = append(fileList, DotsFolderData)
	}

	file, err := os.Open(".")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	sorts.SortWordArr(list)

	for _, name := range list {
		if checks.IsHidden(name) && !skipHidden {
			continue
		}

		var dataToAppend structures.FileData
		AppendData(&dataToAppend, name, saveDirPath)

		if dataToAppend.Name == "" {
			return content
		}

		if dataToAppend.IsDirectory && RFlag {
			subFolderPath := path + "/" + dataToAppend.Name
			dataToAppend.SubFolder = ReadDir(subFolderPath, content, skipHidden, RFlag)
		}

		os.Chdir(saveDirPath)
		fileList = append(fileList, dataToAppend)
		file.Close()
	}

	return fileList
}

// Gets data from all files to save it into array
func AppendData(dataToAppend *structures.FileData, name string, saveDirPath string) {
	fileInfo, err := os.Lstat(name)

	if err != nil {
		return
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		dataToAppend.SymLinkPath, _ = os.Readlink(name)
	}

	timeToAppend := fmt.Sprintf("%+03d:%+03d", fileInfo.ModTime().Hour(), fileInfo.ModTime().Minute())
	timeToAppend = strings.Replace(timeToAppend, "+", "", -1)

	dataToAppend.IsDirectory = fileInfo.IsDir()
	dataToAppend.IsHidden = checks.IsHidden(name)
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

// Gets data from given files in START directory
func DataFromMainDir(files []string, contentList []structures.FileData, flagsToUse structures.Flags, fs *[]structures.FolderContent) {
	if len(files) != 0 {
		var seekingContent []structures.FileData
		sorts.SortWordArr(files)

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

		seekingContent = flags.ApplyFlags(flagsToUse, seekingContent)
		if len(seekingContent) != 0 {
			CollectFiles(seekingContent, seekingContent[0].Path, flagsToUse, fs, true)
		}
	} else {
		CollectFiles(contentList, contentList[0].Path, flagsToUse, fs, false)
	}
}

// Gets data from files in given directory
func DataFromDifferentDir(paths []string, flagsToUse structures.Flags, fs *[]structures.FolderContent) {
	for i := 0; i < len(paths); i++ {
		var tempVar []structures.FileData
		tempVar = ReadDir(paths[i], tempVar, flagsToUse.Flag_a, flagsToUse.Flag_R)
		tempVar = flags.ApplyFlags(flagsToUse, tempVar)
		CollectFiles(tempVar, tempVar[0].Path, flagsToUse, fs, false)
	}
}

// Structures all collected folders into list with files data
func CollectFiles(content []structures.FileData, path string, flagsToUse structures.Flags, res *[]structures.FolderContent, certainFiles bool) {
	var dataToAppend structures.FolderContent
	var totalCalculated bool

	if certainFiles {
		dataToAppend.Path = "null"
	} else {
		dataToAppend.Path = path + ":"
	}

	for i := 0; i < len(content); i++ {
		if !flagsToUse.Flag_a && content[i].IsHidden {
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
					dataToAppend.Total = calculations.CalculateBlocks(content)
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
		}
	}

	*res = append(*res, dataToAppend)

	if flagsToUse.Flag_R {
		for i := 0; i < len(content); i++ {
			if content[i].IsDirectory && content[i].Name != "." && content[i].Name != ".." {
				CollectFiles(content[i].SubFolder, path+"/"+content[i].Name, flagsToUse, res, false)
			}
		}
	}
}

// Prints collected data
func PrintData(fs []structures.FolderContent, flagsToUse structures.Flags) {
	for i := 0; i < len(fs); i++ {
		if len(fs) > 1 && !flagsToUse.Flag_R && fs[i].Path != "null" {
			fmt.Println(fs[i].Path)
		}

		if len(fs[i].FileNames) == 0 {
			if flagsToUse.Flag_R && fs[i].Path != "null" {
				fmt.Println(fs[i].Path)
			}
			if flagsToUse.Flag_l && fs[i].Total != -1 {
				fmt.Println("total:", fs[i].Total)
			}
		}

		for k := 0; k < len(fs[i].FileNames); k++ {
			if flagsToUse.Flag_R && k == 0 && fs[i].Path != "null" {
				fmt.Println(fs[i].Path)
			}
			if flagsToUse.Flag_l && k == 0 && fs[i].Total != -1 {
				fmt.Println("total:", fs[i].Total)
			}
			fmt.Print(fs[i].FileNames[k])
		}

		if len(fs) == 1 && !flagsToUse.Flag_l && len(fs[i].FileNames) != 0 {
			fmt.Println()
			continue
		}

		if len(fs) > 1 && !flagsToUse.Flag_l {
			fmt.Println()
		}

		if i != len(fs)-1 && len(fs[i].FileNames) != 0 {
			fmt.Println()
		}
	}
}

// Gets parent folder path of START directory
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
