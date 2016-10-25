package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func initialisation() {
	var err error
	db, err = sqlx.Open("sqlite3", "randomdect.db")
	if err != nil {
		log.Fatal(err)
	}
	initDB()
}

func initDB() {
	cont, err := exists("randomdect.db")
	if err != nil {
		log.Fatal(err)
	}
	if cont {
		fmt.Println("cont")
		return
	}
	_, err = os.Create("randomdect.db")
	if err != nil {
		log.Fatal("Could not create file randomdect.db", err)
	}
	_, err = db.Exec(createstmt)
	if err != nil {
		log.Fatalf("%q: %s\n", err, createstmt)
	}
}

type entry struct {
	Extension int    `xml:"extension"`
	Name      string `xml:"name"`
	Called    int
	Optout    bool
}

type phonebook struct {
	Event   string  `xml:"event"`
	Entries []entry `xml:"entries>entry"`
}

func loadPhonebook(filename string) (phonebook, error) {
	var p phonebook
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return p, err
	}
	err = xml.Unmarshal(body, &p)
	return p, err
}

func (p *phonebook) Insert() error {
	for _, e := range p.Entries {
		var id int
		err := db.Get(&id, "Select Count(*) from Entries Where Extension = ?", e.Extension)
		if err != nil {
			return err
		}
		if id == 0 {
			_, err := db.Exec("Insert Into Entries (extension, name) Values (?,?)", e.Extension, e.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getEntries(s int) ([]entry, error) {
	var ee []entry
	var err error
	switch s {
	case 0:
		err = db.Select(&ee, "Select * from Entries")
	case 1:
		err = db.Select(&ee, "Select * from Entries Where optout = ?", false)
	case 2:
		err = db.Select(&ee, "Select * from Entries Where Optout = ?", true)
	}
	return ee, err
}

func optout(e int) error {
	_, err := db.Exec("Update Entries Set optout=true Where extension=?", e)
	return err
}

func increasecalled(e int) error {
	var et entry
	err := db.Get(&et, "Select * from Entries Where extension = ?", e)
	if err != nil {
		return err
	}
	et.Called = et.Called + 1
	_, err = db.Exec("Update Entries Set called = ? Where extension = ?", et.Called, et.Extension)
	return err
}

func remove(e int) error {
	_, err := db.Exec("Delete from Entries Where extension = ?", e)
	return err
}

const createstmt = `
--
-- File generated with SQLiteStudio v3.0.7 on Mo. Okt. 24 20:54:43 2016
--
-- Text encoding used: UTF-8
--
PRAGMA foreign_keys = off;
BEGIN TRANSACTION;

-- Table: Entries
CREATE TABLE Entries (extension INTEGER PRIMARY KEY NOT NULL, name STRING NOT NULL, called INTEGER NOT NULL DEFAULT (0), optout INTEGER NOT NULL DEFAULT 0);

COMMIT TRANSACTION;
PRAGMA foreign_keys = on;
`
