package main

import (
	"io/ioutil"
	"log"
	"os"
)

func main() {
	dbFileName := ""
	f, err := os.Open(dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	result, err := parse(bts)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatalln(result)
}

type DB struct {
}

func parse(bts []byte) (*DB, error) {
	return nil, nil
}
