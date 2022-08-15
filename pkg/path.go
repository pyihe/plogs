package pkg

import (
	"path"

	"github.com/pyihe/go-pkg/files"
)

func lastChar(p string) uint8 {
	if p == "" {
		panic("The length of the string can't be 0")
	}
	return p[len(p)-1]
}

func JoinPathName(filePath, fileName string) string {
	finalName := path.Join(filePath, fileName)
	if lastChar(finalName) == '/' {
		finalName = fileName[:len(finalName)-1]
	}
	return finalName
}

func JoinPath(p1, p2 string) string {
	if p2 == "" {
		return p1
	}

	finalPath := path.Join(p1, p2)
	if lastChar(p2) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func MakeDir(dir string) error {
	return files.NewPath(dir)
}
