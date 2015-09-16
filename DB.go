package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	DbUser = "dbuser"
	//DB_PASSWORD = ""
	DbName = "testdb"
)

type DB struct {
	connection *sql.DB
}

func (db *DB) connect() {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable",
		DbUser, DbName)

	var err error

	db.connection, err = sql.Open("postgres", dbinfo)
	checkErr(err)

	log.Printf("Connected to database %s", DbName)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

var MyDB = DB{}
