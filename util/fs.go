package util

import (
	"os"
	"strings"
)

func FindRootPath(rootPath string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}

	directories := strings.Split(wd, "/")
	var idx int
	for i, dir := range directories {
		if dir == rootPath {
			idx = i
		}
	}
	rootPaths := directories[:idx+1]
	return strings.Join(rootPaths, "/")
}
