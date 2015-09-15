package main

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
	"log"
)

const (
    DB_USER     = "dbuser"
    //DB_PASSWORD = ""
    DB_NAME     = "testdb"
)

type DB struct {
	connection *sql.DB
}

func (db *DB) connect() {
	dbinfo := fmt.Sprintf("user=%s dbname=%s sslmode=disable",
        DB_USER, DB_NAME)

	var err error

	db.connection, err = sql.Open("postgres", dbinfo)
    checkErr(err)

	log.Printf("Connected to database %s", DB_NAME)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

var MyDB = DB {

}
