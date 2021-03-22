package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/rbxb/httpfilter"
	"github.com/rbxb/ipskimmer"
)

var port string
var root string
var logPath string

func init() {
	flag.StringVar(&port, "port", ":8080", "The address and port the fileserver listens at.")
	flag.StringVar(&root, "root", "./root", "The directory serving files.")
	flag.StringVar(&logPath, "log", "", "The log file to write to.")
}

func main() {
	flag.Parse()
	if logPath != "" {
		f, err := os.OpenFile("ipskimmer.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
	sk := ipskimmer.NewServer(root)
	fs := httpfilter.NewServer(root, "", map[string]httpfilter.OpFunc{
		"sk": func(w http.ResponseWriter, req *http.Request, args ...string) {
			sk.ServeHTTP(w, req)
		},
	})
	if err := http.ListenAndServe(port, fs); err != nil {
		log.Fatal(err)
	}
}
