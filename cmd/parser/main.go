package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const (
	pageSize = 4 << 10
)

func main() {
	dbFileName := "main.db"
	f, err := os.Open(dbFileName)
	if err != nil {
		log.Fatal(err)
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	db, err := parse(bts)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatalln(db.String())
}

type DBInfo struct {
	Mp [2]MetaPage
}

func (di *DBInfo) Read(bts []byte) (int, error) {
	_, err := di.Mp[0].Read(bts[:pageSize])
	if err != nil {
		return 0, err
	}

	_, err = di.Mp[1].Read(bts[pageSize:])
	if err != nil {
		return 0, err
	}

	return 2 * pageSize, nil
}

func (di *DBInfo) String() string {
	bts, err := json.MarshalIndent(di, "", " ")
	if err != nil {
		return "error json"
	}

	return string(bts)
}

type MetaPage struct {
	Page page
}

func (mp *MetaPage) Read(bts []byte) (int, error) {
	if len(bts) < 16 {
		return 0, fmt.Errorf("len(bts) = %d", len(bts))
	}
	mp.Page.Id = binary.LittleEndian.Uint64(bts[:8])
	mp.Page.Flags = binary.LittleEndian.Uint16(bts[8:10])
	mp.Page.Count = binary.LittleEndian.Uint16(bts[10:12])
	mp.Page.Overflow = binary.LittleEndian.Uint32(bts[12:16])
	return 16, nil
}

func (mp *MetaPage) String() string {
	bts, err := json.Marshal(mp)
	if err != nil {
		return "error MetaPage"
	}

	return string(bts)
}

type page struct {
	Id       uint64 // 页id
	Flags    uint16
	Count    uint16
	Overflow uint32
}

func parse(bts []byte) (*DBInfo, error) {
	var db = &DBInfo{}
	_, err := db.Read(bts)
	if err != nil {
		return nil, err
	}
	return db, nil
}
