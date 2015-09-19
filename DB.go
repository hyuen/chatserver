package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

const (
	//	DbUser = "dbuser"
	//DB_PASSWORD = ""
	//	DbName = "testdb"

	DbUser     = "root"
	DbPassword = "SCUubLc8dhRS2Qt4"
	DbName     = "db"
)

type DB struct {
	connection *sql.DB
}

func (db *DB) connect() {
	dbinfo := os.Getenv("DATABASE_URL")
	//fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable",
	//	DbUser, DbName, DbPassword)

	var err error

	db.connection, err = sql.Open("postgres", dbinfo)
	checkErr(err)
	err = db.connection.Ping()
	checkErr(err)

	log.Info("Connected to database %s", DbName)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

var MyDB = DB{}
