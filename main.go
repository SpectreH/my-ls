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

var STARTDIR string
var USERHOMEDIR string

func ReadDir(path string) {
	// var pathToRead string
	// var arr []FileData
	homeDir, _ := os.UserHomeDir()

	if strings.Contains(path, homeDir) {
		os.Chdir(homeDir)
	} else {
		os.Chdir(STARTDIR)
	}

	file, err := os.Open("/mnt/c/Users/thega/GoProjects/my-ls/test")
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

		if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
			UID, _ := user.LookupId(strconv.Itoa(int(stat.Uid)))
			GID, _ := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))

			dataToAppend.Hardlinks = int(stat.Nlink)
			dataToAppend.SizeKB = int(stat.Blocks / 2)
			dataToAppend.Owner = UID.Username
			dataToAppend.Group = GID.Name
		}

		fmt.Println(dataToAppend)
	}
}

func isHidden(filename string) bool {
	if filename[0:1] == "." {
		return true
	}

	return false
}

func CollectFlagsAndPaths(arguments []string) (Flags, []string) {
	var flagsToUse Flags
	var paths []string

	inputArgs := arguments[1:]

	for i := 0; i < len(inputArgs); i++ {
		if inputArgs[i] == "-" {
			paths = append(paths, inputArgs[i])
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

	return flagsToUse, paths
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
	STARTDIR, _ = os.Getwd()
	flagsToUse, paths := CollectFlagsAndPaths(os.Args)

	if len(paths) == 0 {
		ReadDir(".")
	} else {
		for i := 0; i < len(paths); i++ {
			ReadDir(paths[i])
		}
	}

	fmt.Println(flagsToUse, paths)
}
