package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

    _ "github.com/lib/pq"
	"github.com/gorilla/mux"
)

type User struct {
    Id int `json:"id"` 
    Name string `json:"name"` 
    Email int `json:"email"` 
}

func main() {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // create table if it doesnt eist
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
    if err != nil {
        log.Fatal(err)
    }

    // router
    router := mux.NewRouter()
    router.HandleFunc("/api/go/users", getUsers(db)).Methods("GET")
    router.HandleFunc("/api/go/users", createUser(db)).Methods("POST")
    router.HandleFunc("/api/go/users/{id}", getUser(db)).Methods("GET")
    router.HandleFunc("/api/go/users/{id}", updateUser(db)).Methods("PUT")
    router.HandleFunc("/api/go/users/{id}", deleteUser(db)).Methods("DELETE")

    // wrap the router with CORS and JSON content type middleware
    enhancedRouter := enableCORS(jsonContentTypeMiddleware(router))

    log.Fatal(http.ListenAndServe(":8000", enhancedRouter))
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // set cors headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // allow any origin
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // allow any origin
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // allow any origin

        // check if the request is for CORS preflight
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        // pass request to middleware
        next.ServeHTTP(w, r)
    })
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

// Get all users
func getUsers(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    users := []User{} // array of users
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil {
            log.Fatal(err)
        }
        users = append(users, u)
    }
    if err := rows.Err(); err != nil {
        log.Fatal(err)
    }

    json.NewEncoder(w).Encode(users)

    }
}

// get user by id 
func getUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]

        var u User
        err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.Id, &u.Name, &u.Email)
            if err != nil {
                w.WriteHeader(http.StatusNotFound)
                return
            }
            
            json.NewEncoder(w).Encode(u)
    }
}

// create user
func createUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var u User
        json.NewDecoder(r.Body).Decode(&u)

        err := db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.Id)
        if err != nil {
            log.Fatal(err)
        }

        json.NewEncoder(w).Encode(u)
    }
}

// update the user 
func updateUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var u User
        json.NewDecoder(r.Body).Decode(&u)

        vars := mux.Vars(r)
        id := vars["id"]

        // exe the query
        _, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", u.Name, u.Email, id)
        if err != nil {
            log.Fatal(err)
        }

        var updatedUser User
        err = db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
        if err != nil {
            log.Fatal(err)
        }

        // send the new user in the response
        json.NewEncoder(w).Encode(updatedUser)
    }
}

// delete user
func deleteUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]

        var u User
        err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.Id, &u.Name, &u.Email)
        if err != nil {
            w.WriteHeader(http.StatusNotFound)
            return
        } else {
            _, err := db.Exec("DELETE FROM users WHERE id = $1", id)
            if err != nil {
                // TODO fix error handling
                w.WriteHeader(http.StatusNotFound)
                return
            }

            json.NewEncoder(w).Encode("user deleted")
        
        }

    }
}
