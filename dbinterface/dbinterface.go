package dbinterface

/*
This interface acts as a wrapper around the database, and handles cookies, key generation, login, logout.
*/

import (
	"database/sql"
	"net/http"
	"time"

	"log"

	"errors"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

import _ "github.com/go-sql-driver/mysql"

type DB struct {
	db *sql.DB                    // actual MySQL hook
	sc *securecookie.SecureCookie // encrypts/decrypts cookies to check for validity
}

/*
When the dbinterface starts up, it generates a random key that will be used to both encrypt and decrypt cookie values.
It works as a basic form of encryption, but it is still symmetric.

It should be pretty crackable assuming someone wants to put in the time, but it's very simple to improve the security here
and the worst case scenario is someone gets to see someone else's apikey, which is not the end of the world.
*/
func NewDB() *DB {
	credentials := "root:@/"
	db, err := sql.Open("mysql", credentials) // assumes there is a MySQL instance existing with user root and no password
	if err != nil {
		panic(err)
	}

	// CREATES DATABASE aicomp IF IT DOESN'T EXIST
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS aicomp;")
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", credentials+"aicomp")
	if err != nil {
		panic(err)
	}

	// CREATE TABLE users WITHIN aicomp IF IT DOESN'T EXIST
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
	createdAt DATETIME, name VARCHAR(50), email VARCHAR(50), username VARCHAR(50), 
	pictureLoc VARCHAR(50), password VARCHAR(120), apikey VARCHAR(50));`)
	if err != nil {
		panic(err)
	}

	return &DB{
		db: db,
		sc: securecookie.New(GenerateKey(true, true, true, true), nil), // uses keygen from same pkg
	}
}

// Closes DB, should be deferred to right before ending things
func (d *DB) Close() {
	d.db.Close()
}

/*
Takes a username and password, and returns valid cookie, error=nil if valid login
Otherwise returns nil, valid error
*/
func (d *DB) VerifyLogin(username, password string) (*http.Cookie, error) {
	var databaseUsername string
	var databasePassword string

	err := d.db.QueryRow("SELECT username, password FROM users WHERE username=?", username).Scan(&databaseUsername, &databasePassword)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	if err != nil {
		return nil, err
	}

	if cookie, err := d.generateLoginCookie(databaseUsername); err == nil {
		return cookie, nil
	}
	return nil, err
}

// create a new cookie with name login and value of encoded username (with secret key generated auto on startup)
// when saving a cookie, it will automatically overwrite the cookie of the same name, so login should be the name always.
func (d *DB) generateLoginCookie(username string) (*http.Cookie, error) {
	if encoded, err := d.sc.Encode("login", username); err == nil {
		expiration := time.Now().Add(365 * 24 * time.Hour) // expires in 1 year
		return &http.Cookie{
			Name:    "login",
			Value:   encoded,
			Path:    "/",
			Expires: expiration,
		}, nil
	} else {
		return nil, err
	}
}

// verifies a login cookie from a http.Request, if valid returns empty string (can't return nil), nil
// otherwise returns nil, valid error
func (d *DB) VerifyCookie(cookie *http.Cookie) (string, error) {
	var username string
	if err := d.sc.Decode("login", cookie.Value, &username); err == nil {
		return username, nil
	} else {
		return "", err
	}
}

// User represents the all of the data stored about a user, less their password
type User struct {
	Name           string
	Email          string
	ProfilePicture string
	Username       string
	Apikey         string
}

// GetUser returns the user information attached to the given username
func (d *DB) GetUser(username string) (*User, error) {
	var name string
	var email string
	var pictureLoc string
	var apikey string
	err := d.db.QueryRow("SELECT name, email, pictureLoc, apikey FROM users WHERE username=?",
		username).Scan(&name, &email, &pictureLoc, &apikey)
	if err != nil {
		return nil, err
	}
	return &User{
		Name:           name,
		Email:          email,
		ProfilePicture: pictureLoc,
		Username:       username,
		Apikey:         apikey,
	}, nil
}

func (d *DB) GetUserFromApiKey(apikey string) (*User, error) {
	var username string
	err := d.db.QueryRow("SELECT username, apikey FROM users WHERE apikey=?", apikey).Scan(&username, &apikey)
	if err != nil {
		return nil, err
	}
	return &User{
		Username: username,
		Apikey:   apikey,
	}, nil
}

// SignupUser creates a new user in the database given a user struct and password. On success,
// 	returns a valid *http.Cookie
// Any sort of verification should be handled by server/server.go or the front-end!
func (d *DB) SignupUser(user *User, password string) (*http.Cookie, error) {
	var username string
	err := d.db.QueryRow("SELECT username FROM users WHERE username=?", user.Username).Scan(&username)
	switch {
	// Username is available, so we create a new user
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		apikey := string(GenerateUniqueKey(true, true, true, true))
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		now := time.Now()
		pictureLoc := ""
		_, err = d.db.Exec(`INSERT INTO users(createdAt, name, email, username, pictureLoc, password, 
apikey) VALUES(?, ?, ?, ?, ?, ?, ?)`, now, user.Name, user.Email, user.Username, pictureLoc,
			hashedPassword, apikey)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		cookie, err := d.generateLoginCookie(user.Username)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		// Successful return
		return cookie, nil
	// Some error occured with the database while processing the query
	case err != nil:
		log.Fatalln(err)
		return nil, err
	// This username already exists
	default:
		err = errors.New("username exists")
		return nil, err
	}
}
