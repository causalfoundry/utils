package util

import (
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

func AppRootPath(folderName string) string {
	wd, err := os.Getwd()
	Panic(err)
	directories := strings.Split(wd, "/")
	rootIdx := slices.Index(directories, folderName)
	rootPaths := directories[:rootIdx+1]
	return strings.Join(rootPaths, "/")
}
