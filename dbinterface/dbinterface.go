package dbinterface

/*
This interface acts as a wrapper around the database, and handles cookies, key generation, login, logout.
*/

import (
	"database/sql"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"

	"bufio" // file-reading imports to deal with CREDENTIALS file
	"os"
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
	var credentials string
	file, err := os.Open("dbinterface/CREDENTIALS")
	if err != nil { // default, if credentials doesn't exist
		credentials = "root:@/" // USAGE: "username:password@IP-address/"
	} else { // TODO: add IP address to CREDENTIALS file?
		defer file.Close()
		scanner := bufio.NewScanner(file)
		scanner.Scan() // get username "user:<username>"
		text := scanner.Text()
		credentials = text[5:]
		scanner.Scan() // get password "pass:<password>"
		text = scanner.Text()
		credentials += ":" + text[5:] + "@/"
	}
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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, username VARCHAR(50), password VARCHAR(120), apikey VARCHAR(50));`)
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

type Profile struct {
	Username string
	Apikey   string
}

// given username, outputs Profile object
func (d *DB) GetProfile(username string) (*Profile, error) {
	var apikey string
	err := d.db.QueryRow("SELECT username, apikey FROM users WHERE username=?", username).Scan(&username, &apikey)
	if err != nil {
		return nil, err
	}
	return &Profile{
		Username: username,
		Apikey:   apikey,
	}, nil
}

func (d *DB) GetProfileFromApiKey(apikey string) (*Profile, error) {
	var username string
	err := d.db.QueryRow("SELECT username, apikey FROM users WHERE apikey=?", apikey).Scan(&username, &apikey)
	if err != nil {
		return nil, err
	}
	return &Profile{
		Username: username,
		Apikey:   apikey,
	}, nil
}

// TODO check to see how it's handled when user tries to add duplicates
// AUTOMATICALLY ADDS USER TO DB
// Any sort of verification should probably be handled by server/server.go or the front-end!!!
func (d *DB) SignupUser(username, password string) (*http.Cookie, error) {
	var user string
	err := d.db.QueryRow("SELECT username FROM users WHERE username=?", username).Scan(&user)
	switch { // Username is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		apikey := string(GenerateUniqueKey(true, true, true, true))
		if err != nil {
			return nil, err
		}

		_, err = d.db.Exec("INSERT INTO users(username, password, apikey) VALUES(?, ?, ?)", username, hashedPassword, apikey)
		if err != nil {
			return nil, err
		}

		cookie, err := d.generateLoginCookie(username)
		if err != nil {
			return nil, err
		}
		return cookie, nil
	case err != nil:
		return nil, err
	default: // not sure what happens if it gets here? maybe this is if you try to add already existing user?
		return nil, err
	}
}
