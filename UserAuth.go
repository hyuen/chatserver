package main

import (
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"fmt"
	"log"
	"crypto/rand"
)

type User struct {
 ID           int
 Username     string
 PasswordHash string
 PasswordSalt string
 IsDisabled   bool
}

type UserSession struct {
 SessionKey   string
 LoginTime    time.Time
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/html")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "authing %s", vars["operation"])
	
}

// See if user is in the system, if yes assign a session ID to him
func UserAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/html")
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	row := MyDB.connection.QueryRow("select id from users where username=$1 and passwordhash=$2",
                             		username, password)

	var user_id int

	err := row.Scan(&user_id)
	if err != nil {
		log.Print("Unknown login or password for ", username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Printf("user_id = %d", user_id)
	b := make([]byte, 24)
	rand.Read(b)
	session_id := fmt.Sprintf("%x", b)
	
	_, err = MyDB.connection.Query("insert into usersessions(sessionkey, user_id) values($1,$2)",
		                         session_id, user_id)

	if err != nil {
		panic(err)
	}
	
	// Set output
	c := http.Cookie { Name: "sessionId", Value: session_id }
	http.SetCookie(w, &c)

	log.Printf("Login successful for user:%s", username)
}
