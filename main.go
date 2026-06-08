package main

import (
	"fmt"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hi tis I")
}

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/login.html")
		return
	}

	roll := r.FormValue("roll")
	password := r.FormValue("password")

	fmt.Println("Login Attempt")
	fmt.Println("Roll:", roll)
	fmt.Println("Password:", password)

	fmt.Fprintln(w, "Logged in")
}

func signup(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "templates/signup.html")
		return
	}

	roll := r.FormValue("roll")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	fmt.Println("Roll:", roll)
	fmt.Println("Name:", name)
	fmt.Println("Email:", email)
	fmt.Println("Password:", password)

	fmt.Fprintln(w, "Signed up")
}

func main() {

	http.HandleFunc("/", home)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)

	fmt.Println("Server running on http://localhost:3000")

	http.ListenAndServe(":3000", nil)
}
