package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jpfairbanks/querygarden/log"
	// "html/template"
	"net/http"
	"os"
)

//Store the cookie store which is going to store session data in the cookie
var Store = sessions.NewCookieStore([]byte(os.Getenv("CORSAIR_SECRET")))

//IsLoggedIn will check if the user has an active session and return True
func IsLoggedIn(r *http.Request) bool {
	session, _ := Store.Get(r, "session")
	if session.Values["loggedin"] == "true" {
		return true
	}
	return false
}

func Login(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		// log.Debug(os.Getenv("CORSAIR_PASSWORD"))
		session, _ := Store.Get(r, "session")
		if r.PostFormValue("logout") == "Logout" {
			log.Infof("Logging out: %s", session.Values["username"])
			session.Values["loggedin"] = "false"
			return session.Save(r, w)
		}
		if r.PostFormValue("password") == os.Getenv("CORSAIR_PASSWORD") {
			session.Values["loggedin"] = "true"
			session.Values["username"] = r.PostFormValue("Username")
			log.Infof("Logged in user: %s", r.PostFormValue("Username"))
			session.Save(r, w)
			return nil
		} else {
			// log.Debugf("Password Supplied: %s", r.PostFormValue("password"))
			log.Infof("Bad username/password: %s",
				r.PostFormValue("Username"))
			return fmt.Errorf("Bad username/password")
		}
	}
	return nil
}

func loginhandler(w http.ResponseWriter, r *http.Request) {
	err := Login(w, r)
	var lerr bool
	lerr = err != nil
	d := map[string]interface{}{
		"loggedin":   IsLoggedIn(r),
		"LoginError": lerr,
		"Message":    err,
	}
	if err != nil {
		d["Message"] = err.Error()
	}
	err = templates.ExecuteTemplate(w, "login.html.tmpl", d)
	if err != nil {
		log.Error(err)
	}
}
