package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rbxb/ipskimmer"
)

var root string

var now = time.Now().Unix()

func init() {
	flag.StringVar(&root, "root", "./root", "The working directory.")
}

func main() {
	if err := filepath.WalkDir(filepath.Join(root, "links"), fs.WalkDirFunc(walk)); err != nil {
		log.Println(err)
	}
}

func walk(path string, info fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	expires, err := ipskimmer.ReadLinkExpires(path)
	if err != nil {
		return err
	}
	if expires < now {
		os.Remove(path)
		os.Remove(filepath.Join(root, "visitors", info.Name()))
		log.Println("Removed ", info.Name())
	}
	return nil
}
