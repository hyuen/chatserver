package main

import (
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"fmt"
	"log"
	
	"utils"
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

	row := MyDB.connection.QueryRow(`select id, passwordhash, passwordsalt, isdisabled
                                     from users where username=$1`,
                             		 username)

	var user_id int
	var passwordhash string
	var passwordsalt string
	var isdisabled bool

	err := row.Scan(&user_id, &passwordhash, &passwordsalt, &isdisabled)
	if err != nil {
		log.Print("Unknown login or password for ", username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if isdisabled {
		log.Printf("Username %s is disabled", username)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	calculated_passwordhash := utils.SHA1(passwordsalt + password)

	if calculated_passwordhash != passwordhash {
		log.Printf("Invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// creating session
	session_id := utils.RandomSHA1()
	for tries := 0; tries < 3; tries++ {
		_, err = MyDB.connection.Exec("insert into usersessions(sessionkey, user_id) values($1,$2)",
			session_id, user_id)

		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}
	
	// Set output
	c := http.Cookie { Name: "sessionId", Value: session_id }
	http.SetCookie(w, &c)

	log.Printf("Login successful for user:%s user_id:%d", username, user_id)
}

// log the user's session out of the system, do not check if the session is valid
func UserAuthLogoutHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Add("Content-Type", "text/html")
	vars := mux.Vars(r)
	
	session_id := vars["session_id"]
	log.Print("logging out ", session_id)

	_, err := MyDB.connection.Exec("delete from usersessions where sessionkey = $1", session_id)
	if err != nil {
		panic(err)
	}
}
