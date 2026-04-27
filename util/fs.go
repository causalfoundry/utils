package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindRootPath(rootPath string) string {
	ret, err := FindRootPathE(rootPath)
	if err != nil {
		panic(err.Error())
	}
	return ret
}

func FindRootPathE(rootPath string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return findRootPathFrom(wd, rootPath)
}

func findRootPathFrom(wd, rootPath string) (string, error) {
	if rootPath == "" {
		return "", fmt.Errorf("root path cannot be empty")
	}

	cleanWD := filepath.Clean(wd)
	volume := filepath.VolumeName(cleanWD)
	trimmed := strings.TrimPrefix(cleanWD, volume)
	directories := strings.Split(trimmed, string(filepath.Separator))
	var idx = -1
	for i, dir := range directories {
		if dir == rootPath {
			idx = i
		}
	}
	if idx < 0 {
		return "", fmt.Errorf("root path %q not found in %q", rootPath, wd)
	}
	rootPaths := directories[:idx+1]
	ret := strings.Join(rootPaths, string(filepath.Separator))
	if ret == "" {
		ret = string(filepath.Separator)
	}
	return volume + ret, nil
}
