// +build fatbin

package main

import (
	"os"
	"path/filepath"
)

func init() {
	path, _ := filepath.Split(os.Args[0])
	resourceBlobFile = filepath.Join(path, "resources.blob")
}

var resourceBlobFile string
