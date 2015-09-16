package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

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
	SessionKey string
	LoginTime  time.Time
}

const (
	AuthSalt = "sydrtvgh"
)

type ErrorMsg struct {
	Msg string
}

func UserAuth(username string, password string) (int, string, bool) {
	row := MyDB.connection.QueryRow(`select id, passwordhash, passwordsalt, isdisabled
                                     from users where username=$1`,
		username)

	var userID int
	var passwordHash string
	var passwordSalt string
	var isDisabled bool

	err := row.Scan(&userID, &passwordHash, &passwordSalt, &isDisabled)
	if err != nil {
		log.Print("Unknown login or password for ", username)
		return -1, "", false
	}

	if isDisabled {
		log.Printf("Username %s is disabled", username)
		return -1, "", false
	}

	calculatedPasswordHash := utils.SHA1(passwordSalt + password)
	//log.Printf("calculated hash %s", calculated_passwordhash)

	if calculatedPasswordHash != passwordHash {
		log.Printf("Invalid password")
		return -1, "", false
	}

	return userID, passwordSalt, true
}

// See if user is in the system, if yes assign a session ID to him
func UserAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, _, ok := UserAuth(username, password)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// creating session
	sessionID := utils.RandomSHA1()
	var err error
	for tries := 0; tries < 3; tries++ {
		_, err = MyDB.connection.Exec("insert into usersessions(sessionkey, user_id) values($1,$2)",
			sessionID, userID)

		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}

	// Set output
	c := http.Cookie{Name: "sessionId", Value: sessionID}
	http.SetCookie(w, &c)

	log.Printf("Login successful for user:%s user_id:%d", username, userID)
}

// log the user's session out of the system, do not check if the session is valid
func UserAuthLogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)

	sessionID := vars["session_id"]
	log.Print("logging out ", sessionID)

	_, err := MyDB.connection.Exec("delete from usersessions where sessionkey = $1", sessionID)
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

	userID, salt, ok := UserAuth(username, password)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
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
		errmsg := ErrorMsg{Msg: "The new password cannot be the same as the current one"}
		str, _ := json.Marshal(errmsg)
		w.Write(str)
		return
	}

	newpasswordhash := utils.SHA1(salt + newpassword)

	// update password
	_, err = MyDB.connection.Exec("UPDATE users SET passwordhash=$1 where id=$2",
		newpasswordhash, userID)
	if err != nil {
		panic(err)
	}

	// expire all sessions
	_, err = MyDB.connection.Exec("DELETE FROM usersessions WHERE user_id=$1",
		userID)
	if err != nil {
		panic(err)
	}
}

type SignupForm struct {
	Username string
	Password string
	// Telephone
}

func (form *SignupForm) validate() bool {
	if form.Username == "" || form.Password == "" {
		return false
	}
	return true
}

func UserAuthSignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	decoder := schema.NewDecoder()
	signupForm := new(SignupForm)
	err = decoder.Decode(signupForm, r.PostForm)

	if err != nil || !signupForm.validate() {
		//log.Print(err, signup_form, signup_form.Username)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check if user exists
	var id int
	err = MyDB.connection.QueryRow("SELECT 1 FROM users WHERE username = $1",
		signupForm.Username).Scan(&id)

	if err == nil {
		log.Print("User ", signupForm.Username, " already exists")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passwordhash := utils.SHA1(AuthSalt + signupForm.Password)
	_, err = MyDB.connection.Exec(`INSERT INTO
                                    users(username, passwordhash, passwordsalt)
                                    VALUES($1,$2,$3)`,
		signupForm.Username, passwordhash, AuthSalt)
	if err != nil {
		panic(err)
	}
	log.Printf("Created user %s", signupForm.Username)
}

func SessionRequired(f http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		f(w, r)
	}
}
