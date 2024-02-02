package util

import (
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

func GetUniqueDirPath(dir string) string {
	wd, err := os.Getwd()
	Panic(err)
	directories := strings.Split(wd, "/")
	rootIdx := slices.Index(directories, dir)
	rootPaths := directories[:rootIdx+1]
	return strings.Join(rootPaths, "/")
}
