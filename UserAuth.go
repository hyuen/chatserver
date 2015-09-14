package main

import (
	"net/http"
	"time"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"log"
	"encoding/json"
	
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

type ErrorMsg struct {
	Msg   string
}

func UserAuth(username string, password string) (int, string, bool) {
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
		return -1, "", false
	}

	if isdisabled {
		log.Printf("Username %s is disabled", username)
		return -1, "", false
	}
	
	calculated_passwordhash := utils.SHA1(passwordsalt + password)
	//log.Printf("calculated hash %s", calculated_passwordhash)
	
	if calculated_passwordhash != passwordhash {
		log.Printf("Invalid password")
		return -1, "", false
	}
	
	return user_id, passwordsalt, true
}

// See if user is in the system, if yes assign a session ID to him
func UserAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user_id, _, ok := UserAuth(username, password)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// creating session
	session_id := utils.RandomSHA1()
	var err error
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
    w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	
	session_id := vars["session_id"]
	log.Print("logging out ", session_id)

	_, err := MyDB.connection.Exec("delete from usersessions where sessionkey = $1", session_id)
	if err != nil {
		panic(err)
	}
}

type PasswordChangeForm struct {
	New_password string
}

func UserAuthPasswordChangeHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	
	user_id, salt, ok := UserAuth(username, password)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
		return
	}

	decoder := schema.NewDecoder()
	passwordchange_form := new(PasswordChangeForm)
	err = decoder.Decode(passwordchange_form, r.PostForm)
		
	if err != nil || passwordchange_form.New_password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newpassword := passwordchange_form.New_password

	if newpassword == password {
		errmsg := ErrorMsg { Msg: "The new password cannot be the same as the current one" }
		str, _ := json.Marshal(errmsg)
		w.Write(str)
		return
	}

	newpasswordhash :=  utils.SHA1(salt + newpassword)
	
	// update password
	_, err = MyDB.connection.Exec("UPDATE users SET passwordhash=$1 where id=$2",
		                          newpasswordhash, user_id)
	if err != nil {
		panic(err)
	}
	
	// expire all sessions
	_, err = MyDB.connection.Exec("DELETE FROM usersessions WHERE user_id=$1",
			                      user_id)
	if err != nil {
		panic(err)
	}
}


func UserAuthSignupHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
	//vars := mux.Vars(r)

}
