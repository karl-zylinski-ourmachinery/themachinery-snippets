package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/jlaffaye/ftp"
)

func main() {
	var password string
	var lib string

	flag.StringVar(&password, "password", "", "ftp password")
	flag.StringVar(&lib, "lib", "", "lib zip file")
	flag.Parse()

	if password == "" {
		log.Fatal("No password specified")
	}

	if lib == "" {
		log.Fatal("No library specified")
	}

	c, err := ftp.Dial("160.153.16.15:21")
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login("ourmachinery", password)
	if err != nil {
		log.Fatal(err)
	}

	err = c.ChangeDir("public_html/lib")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(lib)
	if err != nil {
		log.Fatal(err)
	}
	libBase := path.Base(lib)
	err = c.Stor(libBase, f)

	c.Quit()
}
