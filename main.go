package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
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
	Month string
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
	Path     string
	FileName []string
}

var STARTDIR string
var USERHOMEDIR string

func ReadDir(path string, content []FolderContent) ([]FileData, []FolderContent) {
	var fileList []FileData
	var saveDirPath string

	// var dot FileData
	// var dotDot FileData
	// dot.Name = "."
	// dot.isHidden = true
	// dotDot.isHidden = true
	// dotDot.Name = ".."
	// fileList = append(fileList, dot)
	// fileList = append(fileList, dotDot)

	os.Chdir(path)
	saveDirPath = path

	file, err := os.Open(".")
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders

	for _, name := range list {
		fileInfo, _ := os.Stat(name)
		var dataToAppend FileData

		dataToAppend.isDirectory = fileInfo.IsDir()
		dataToAppend.isHidden = isHidden(name)
		dataToAppend.Permission = fmt.Sprintf("%v", fileInfo.Mode())
		dataToAppend.Name = fileInfo.Name()
		dataToAppend.Size = fileInfo.Size()
		dataToAppend.ModificationTime.Day = fileInfo.ModTime().Day()
		dataToAppend.ModificationTime.Month = fileInfo.ModTime().UTC().Format("Jan")
		dataToAppend.ModificationTime.Time = strconv.Itoa(fileInfo.ModTime().Hour()) + ":" + strconv.Itoa(fileInfo.ModTime().Minute())
		dataToAppend.Path = saveDirPath

		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			UID, _ := user.LookupId(strconv.Itoa(int(stat.Uid)))
			GID, _ := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

			dataToAppend.Hardlinks = int(stat.Nlink)
			dataToAppend.SizeKB = int(stat.Blocks / 2)
			dataToAppend.Owner = UID.Username
			dataToAppend.Group = GID.Name
		}

		if dataToAppend.isDirectory {
			subFolderPath := path + "/" + dataToAppend.Name
			dataToAppend.SubFolder, content = ReadDir(subFolderPath, content)
		}

		os.Chdir(saveDirPath)

		fileList = append(fileList, dataToAppend)
	}

	if len(fileList) != 0 {
		var contentToAppend FolderContent
		contentToAppend.Path = fileList[0].Path
		for i := 0; i < len(fileList); i++ {
			contentToAppend.FileName = append(contentToAppend.FileName, fileList[i].Name)
		}
		content = append(content, contentToAppend)
	}

	return fileList, content
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

	fileList, contentList := ReadDir(STARTDIR, contentList)

	fmt.Println(flagsToUse, paths, files)
	fmt.Println(fileList)
	PrintFolders(contentList)
}

func PrintFolders(list []FolderContent) {
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i].Path + ":")
		for k := 0; k < len(list[i].FileName); k++ {
			fmt.Print(list[i].FileName[k])
			fmt.Print(" ")
		}
		fmt.Println("\n")
	}
}
