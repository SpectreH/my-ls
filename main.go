package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
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
}

type Date struct {
	Month string
	Day   int
	Time  string
}

func readCurrentDir() {
	//var arr []FileData

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

func main() {
	readCurrentDir()
}
