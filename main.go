package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connectionURL := os.Getenv("UZORG_DB_URL")

	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		log.Fatal("Could not open postgress connection: ", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Could not ping postgress: ", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		user_id UUID PRIMARY KEY,
		first_name TEXT,
		last_name TEXT,
		email TEXT UNIQUE,
		phone TEXT,
		password TEXT
	)`)
	if err != nil {
		log.Fatal("Could not create users table: ", err)
	}

	// Updated org table without user_id
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orgs (
		org_id UUID PRIMARY KEY,
		name TEXT,
		description TEXT
	)`)
	if err != nil {
		log.Fatal("Could not create orgs table: ", err)
	}

	// New org_users join table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS org_users (
		org_id UUID,
		user_id UUID,
		PRIMARY KEY (org_id, user_id),
		FOREIGN KEY (org_id) REFERENCES orgs(org_id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
	)`)
	if err != nil {
		log.Fatal("Could not create org_users table: ", err)
	}

	upgs := UzorgPgStorer{db: db}
	reqHandler := ReqHandler{uzorgStore: &upgs}

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the UZORG Web Server!"))
	})

	r.Handle("/auth/register", CMW(http.HandlerFunc(reqHandler.registerUser), LoggingMiddleware)).Methods("POST")
	r.Handle("/auth/login", CMW(http.HandlerFunc(reqHandler.Login), LoggingMiddleware)).Methods("POST")

	r.Handle("/api/users/{id}", CMW(http.HandlerFunc(reqHandler.GetUser), LoggingMiddleware, AuthMiddleware)).Methods("GET")
	// add the new handlers
	r.Handle("/api/organizations", CMW(http.HandlerFunc(reqHandler.CreateOrg), LoggingMiddleware, AuthMiddleware)).Methods("POST")
	r.Handle("/api/organizations", CMW(http.HandlerFunc(reqHandler.GetOrgs), LoggingMiddleware, AuthMiddleware)).Methods("GET")
	r.Handle("/api/organizations/{id}", CMW(http.HandlerFunc(reqHandler.GetOrg), LoggingMiddleware, AuthMiddleware)).Methods("GET")
	r.Handle("/api/organizations/{id}/users", CMW(http.HandlerFunc(reqHandler.GetOrgUsers), LoggingMiddleware, AuthMiddleware)).Methods("GET")
	r.Handle("/api/organizations/{id}/users", CMW(http.HandlerFunc(reqHandler.AddUserToOrg), LoggingMiddleware, AuthMiddleware)).Methods("POST")

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
