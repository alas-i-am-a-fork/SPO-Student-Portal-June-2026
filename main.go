package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hi tis I")
}

func promoteAdmin(w http.ResponseWriter, r *http.Request) {

	roll := r.FormValue("roll")
	domain := r.FormValue("domain")

	if strings.TrimSpace(domain) == "" {
		fmt.Fprintln(w, "Domain cannot be empty")
		return
	}

	_, err := db.Exec(
		`UPDATE records
		 SET role='admin',
		     domain=$1
		 WHERE roll=$2`,
		domain,
		roll,
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	http.Redirect(
		w,
		r,
		"/master-admin-home/manage-admins",
		http.StatusSeeOther,
	)
}

func getCurrentUser(r *http.Request) (string, string, sql.NullString, error) {

	cookie, err := r.Cookie("user_roll")
	if err != nil {
		return "", "", sql.NullString{}, err
	}

	var role string
	var domain sql.NullString

	err = db.QueryRow(
		`SELECT role, domain
		 FROM records
		 WHERE roll = $1`,
		cookie.Value,
	).Scan(&role, &domain)

	if err != nil {
		return "", "", sql.NullString{}, err
	}

	return cookie.Value, role, domain, nil
}

func studentHome(w http.ResponseWriter, r *http.Request) {

	_, role, _, err := getCurrentUser(r)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if role != "student" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "templates/student-home.html")
}

func adminHome(w http.ResponseWriter, r *http.Request) {

	_, role, domain, err := getCurrentUser(r)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if role != "admin" || !domain.Valid {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "templates/domain-admin-home.html")
}

func masterHome(w http.ResponseWriter, r *http.Request) {

	_, role, domain, err := getCurrentUser(r)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if role != "admin" || domain.Valid {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, "templates/master-admin-home.html")
}

type Admin struct {
	Roll   int64
	Name   string
	Domain string
}

type Student struct {
	Roll int64
	Name string
}

type ManageAdminsPageData struct {
	Message  string
	Admins   []Admin
	Students []Student
}

func renderManageAdmins(
	w http.ResponseWriter,
	message string,
	admins []Admin,
	students []Student,
) {

	tmpl := template.Must(
		template.ParseFiles("templates/manage-admins.html"),
	)

	tmpl.Execute(w, ManageAdminsPageData{
		Message:  message,
		Admins:   admins,
		Students: students,
	})
}

func manageAdmins(w http.ResponseWriter, r *http.Request) {

	_, role, domain, err := getCurrentUser(r)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if role != "admin" || domain.Valid {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var admins []Admin
	var students []Student

	// Fetch current domain admins
	adminRows, err := db.Query(
		`SELECT roll, name, domain
		 FROM records
		 WHERE role = 'admin'
		 AND domain IS NOT NULL`,
	)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer adminRows.Close()

	for adminRows.Next() {

		var admin Admin

		err := adminRows.Scan(
			&admin.Roll,
			&admin.Name,
			&admin.Domain,
		)

		if err != nil {
			fmt.Println(err)
			return
		}

		admins = append(admins, admin)
	}

	// Fetch students eligible for promotion
	studentRows, err := db.Query(
		`SELECT roll, name
		 FROM records
		 WHERE role = 'student'`,
	)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer studentRows.Close()

	for studentRows.Next() {

		var student Student

		err := studentRows.Scan(
			&student.Roll,
			&student.Name,
		)

		if err != nil {
			fmt.Println(err)
			return
		}

		students = append(students, student)
	}

	renderManageAdmins(
		w,
		"",
		admins,
		students,
	)
}

func logout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:   "user_roll",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

type LoginPageData struct {
	Message string
}

func renderLogin(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("templates/login.html"))

	tmpl.Execute(w, LoginPageData{
		Message: message,
	})
}

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		renderLogin(w, "")
		return
	}

	roll := r.FormValue("roll")
	password := r.FormValue("password")

	var storedPassword string
	var role string
	var domain sql.NullString

	err := db.QueryRow(
		`SELECT password, role, domain
	 FROM records
	 WHERE roll = $1`,
		roll,
	).Scan(&storedPassword, &role, &domain)

	if err != nil {
		renderLogin(w, "Account not found. Sign up?")
		return
	}

	if password != storedPassword {
		renderLogin(w, "Incorrect password")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_roll",
		Value:    roll,
		Path:     "/",
		HttpOnly: true,
	})

	if role == "student" {
		http.Redirect(w, r, "/student-home", http.StatusSeeOther)
		return
	}

	if role == "admin" {

		if !domain.Valid {
			http.Redirect(w, r, "/master-admin-home", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/domain-admin-home", http.StatusSeeOther)
		return
	}
}

type SignupPageData struct {
	Message string
}

func renderSignup(w http.ResponseWriter, message string) {
	tmpl := template.Must(template.ParseFiles("templates/signup.html"))

	tmpl.Execute(w, SignupPageData{
		Message: message,
	})
}

func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		renderSignup(w, "")
		return
	}

	roll := r.FormValue("roll")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	var count int

	err := db.QueryRow(
		"SELECT COUNT(*) FROM records WHERE roll = $1",
		roll,
	).Scan(&count)

	if err != nil {
		fmt.Println(err)
		return
	}

	if count > 0 {
		renderSignup(w, "Account already exists. Log in?")
		return
	}

	_, err = db.Exec(
		`INSERT INTO records
	(roll, name, email, password, role)
	VALUES ($1, $2, $3, $4, $5)`,
		roll,
		name,
		email,
		password,
		"student",
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	renderSignup(w, "Account created. You may log in.")
}

func main() {
	connectDB()

	http.HandleFunc("/", home)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/student-home", studentHome)
	http.HandleFunc("/domain-admin-home", adminHome)
	http.HandleFunc("/master-admin-home", masterHome)
	http.HandleFunc("/master-admin-home/manage-admins", manageAdmins)
	http.HandleFunc("/logout", logout)

	fmt.Println("Server running on http://localhost:3000")

	http.ListenAndServe(":3000", nil)
}
