package user

import (
	"net/http"
	"onlineshop/helper"
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Login handler
func Login(w http.ResponseWriter, r *http.Request) {
	var ctx interface{}
	var tplPath = map[string]string{
		"folder": "auth",
		"file":   "login.gohtml",
	}

	if r.Method == http.MethodGet {
		result, err := sessionExists(r)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		if result {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else if r.Method == http.MethodPost {
		user := &User{
			Email: r.PostFormValue("email"),
		}
		password := r.PostFormValue("password")

		// Form validation
		vld := &Validator{
			User:   user,
			Errors: map[string]string{},
		}
		match := regexp.MustCompile(".+@.+\\..+").Match([]byte(user.Email))
		if match == false {
			vld.Errors["Email"] = "Please enter a valid email address"
			helper.RenderTemplate(w, tplPath, vld)
			return
		}

		result, err := user.exists()
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		if result == false {
			vld.Errors["Email"] = "User not found with this email address"
			helper.RenderTemplate(w, tplPath, vld)
			return
		}

		// Does the entered password match the stored password?
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			user.Password = password
			vld.Errors["Password"] = "Password does not match, please try again"
			helper.RenderTemplate(w, tplPath, vld)
			return
		}

		// Generate a new uuid && create a new session in the db
		sessionID, _ := uuid.NewV4()
		err = createSession(sessionID.String(), user.ID)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// Create a new session in the browser cookie
		cookie := &http.Cookie{
			Name:  "session_id",
			Value: sessionID.String(),
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
	}

	helper.RenderTemplate(w, tplPath, ctx)
}

// Register handler
func Register(w http.ResponseWriter, r *http.Request) {
	var ctx interface{}
	var tplPath = map[string]string{
		"folder": "auth",
		"file":   "register.gohtml",
	}

	if r.Method == http.MethodGet {
		// Check if session already exists
		result, err := sessionExists(r)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		if result {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	} else if r.Method == http.MethodPost {
		user := &User{
			FirstName: r.PostFormValue("first_name"),
			LastName:  r.PostFormValue("last_name"),
			Email:     r.PostFormValue("email"),
			Password:  r.PostFormValue("password"),
		}
		passwordConfirm := r.PostFormValue("password_confirm")

		// Form validation
		vld := &Validator{User: user}
		if vld.validate(passwordConfirm) == false {
			helper.RenderTemplate(w, tplPath, vld)
			return
		}

		// Check if user already exists
		result, err := user.exists()
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		if result {
			vld.Errors["Email"] = "The email address is already taken. Please choose another one"
			helper.RenderTemplate(w, tplPath, vld)
			return
		}

		// Encrypt password with bcrypt
		passwordSlice, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		user.Password = string(passwordSlice)

		// Create a new user
		userID, err := user.createUser()
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// Generate a new uuid && create a new session in the db
		sessionID, _ := uuid.NewV4()
		err = createSession(sessionID.String(), userID)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// Create a new session in the browser cookie
		cookie := &http.Cookie{
			Name:  "session_id",
			Value: sessionID.String(),
		}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	} else {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	helper.RenderTemplate(w, tplPath, ctx)
}

// Logout handler
func Logout(w http.ResponseWriter, r *http.Request) {
	result, err := sessionExists(r)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	if result == false {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Delete the session from db
	cookie, _ := r.Cookie("session_id")
	err = deleteSession(cookie.Value)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	// Remove the cookie from browser
	cookie = &http.Cookie{
		Name:   "session_id",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (vld *Validator) validate(passwordConfirm string) bool {
	vld.Errors = map[string]string{}
	firstName := strings.TrimSpace(vld.User.FirstName)
	lastName := strings.TrimSpace(vld.User.LastName)
	email := strings.TrimSpace(vld.User.Email)
	password := strings.TrimSpace(vld.User.Password)
	passwordConfirm = strings.TrimSpace(passwordConfirm)

	if firstName == "" || len(firstName) > 20 {
		vld.Errors["FirstName"] = "The field First Name must be a string with a maximum length of 20"
	}

	if lastName == "" || len(lastName) > 20 {
		vld.Errors["LastName"] = "The field Last Name must be a string with a maximum length of 20"
	}

	match := regexp.MustCompile(".+@.+\\..+").Match([]byte(email))
	if match == false {
		vld.Errors["Email"] = "Please enter a valid email address"
	}

	if len(password) < 6 || len(password) > 20 {
		vld.Errors["Password"] = "Your password must be 6-12 characters long"
	}

	if password != passwordConfirm {
		vld.Errors["PasswordConfirm"] = "The specified passwords do not match"
	}

	return len(vld.Errors) == 0
}
