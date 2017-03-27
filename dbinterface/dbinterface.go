package dbinterface

/*
This interface acts as a wrapper around the database, and handles cookies, key generation, login, logout.
*/

import (
	"database/sql"
	"math"
	"net/http"
	"time"

	// file-reading imports to deal with CREDENTIALS file
	"log"

	"errors"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

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
func makeCred(dbuser, dbpass, dbaddr string) string {
	return dbuser + ":" + dbpass + "@" + dbaddr + "/"
}
func NewDB(dbaddr, dbname, dbuser, dbpass string) *DB {
	creds := makeCred(dbuser, dbpass, dbaddr)
	db, err := sql.Open("mysql", creds)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbname + ";")
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("mysql", creds+dbname)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users(id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
	createdAt DATETIME, name VARCHAR(50), email VARCHAR(50), username VARCHAR(50), ELO FLOAT,
	pictureLoc VARCHAR(50), password VARCHAR(120), apikey VARCHAR(50));`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS replays(id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
createdAt DATETIME, gameName VARCHAR(100), fullReplayName VARCHAR(150), username1 VARCHAR(50), username2 VARCHAR(50));`)
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
	ELO            float64
	Apikey         string
}

// GetUser returns the user information attached to the given username
func (d *DB) GetUser(username string) (*User, error) {
	var name string
	var email string
	var ELO float64
	var pictureLoc string
	var apikey string
	err := d.db.QueryRow("SELECT name, email, ELO, pictureLoc, apikey FROM users WHERE username=?",
		username).Scan(&name, &email, &ELO, &pictureLoc, &apikey)
	if err != nil {
		return nil, err
	}
	return &User{
		Name:           name,
		Email:          email,
		ProfilePicture: pictureLoc,
		Username:       username,
		ELO:            ELO,
		Apikey:         apikey,
	}, nil
}

// GetUserList gets all users, pretty much exclusively for display
func (d *DB) GetLeaderboard() ([]User, error) {
	users := []User{}
	rows, err := d.db.Query("SELECT name, email, ELO, pictureLoc, username FROM users ORDER BY ELO DESC")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name string
		var email string
		var elo float64
		var pictureLoc string
		var username string
		err = rows.Scan(&name, &email, &elo, &pictureLoc, &username)
		if err != nil {
			return nil, err
		}
		users = append(users, User{
			Name:           name,
			Email:          email,
			ProfilePicture: pictureLoc,
			Username:       username,
			ELO:            elo,
		})
	}
	return users, nil
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
	// check if the username already exists
	var username string
	err := d.db.QueryRow("SELECT username FROM users WHERE username=?", user.Username).Scan(&username)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalln(err)
		return nil, err
	}
	if username != "" {
		err = errors.New("username exists")
		return nil, err
	}
	// check if the email already exists
	var email string
	err = d.db.QueryRow("SELECT email FROM users WHERE email=?", user.Email).Scan(&email)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalln(err)
		return nil, err
	}
	if email != "" {
		err = errors.New("email exists")
		return nil, err
	}
	// no problems, so go ahead and create the new user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	apikey := string(GenerateUniqueKey(true, true, true, true))
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	now := time.Now()
	_, err = d.db.Exec(`INSERT INTO users(createdAt, name, email, username, pictureLoc, password, 
apikey, ELO) VALUES(?, ?, ?, ?, ?, ?, ?, 1500.0)`, now, user.Name, user.Email, user.Username,
		user.ProfilePicture, hashedPassword, apikey)
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
}

// UpdateELO updates each user's ELO score based on the results from a game.
// winner = 0 if tie
// 1 if username1 is the winner
// 2 if username2 is the winner
func (d *DB) UpdateELO(username1, username2 string, winner int) (*float64, *float64, error) {
	const K = 64.0
	var elo1 float64
	var elo2 float64
	err := d.db.QueryRow("SELECT ELO FROM users WHERE username=?", username1).Scan(&elo1)
	if err != nil {
		return nil, nil, err
	}
	err = d.db.QueryRow("SELECT ELO FROM users WHERE username=?", username2).Scan(&elo2)
	if err != nil {
		return nil, nil, err
	}
	R1 := math.Pow(10, elo1/400)
	R2 := math.Pow(10, elo2/400)
	E1 := R1 / (R1 + R2)
	E2 := R2 / (R1 + R2)
	var S1 float64
	var S2 float64
	if winner == 2 {
		S1 = 0.0
		S2 = 1.0
	} else if winner == 1 {
		S1 = 1.0
		S2 = 0.0
	} else {
		S1 = 0.5
		S2 = 0.5
	}
	newELO1 := elo1 + K*(S1-E1)
	newELO2 := elo2 + K*(S2-E2)
	_, err = d.db.Exec("UPDATE users SET ELO=? WHERE username=?", newELO1, username1)
	if err != nil {
		return nil, nil, err
	}
	_, err = d.db.Exec("UPDATE users SET ELO=? WHERE username=?", newELO2, username2)
	if err != nil {
		return nil, nil, err
	}
	return &newELO1, &newELO2, nil
}

/*
REPLAY RELATED FUNCTIONALITY
*/

// adding a new game into the replay database
func (d *DB) AddGame(replayname, gamename string) {
	_, err := d.db.Exec(`INSERT INTO replays(createdAt, gameName, fullReplayName, username1, username2) VALUES(?, ?, ?, ?, ?)`,
		time.Now(), gamename, replayname, "", "")
	if err != nil {
		log.Fatalln(err)
	}
}

// adding the first player to a game (important for replays)
func (d *DB) AddPlayerOne(replayName string, username string) {
	_, err := d.db.Exec("UPDATE replays SET username1=? WHERE fullReplayName=?", username, replayName)
	if err != nil {
		panic(err)
	}
}

// same with second
func (d *DB) AddPlayerTwo(replayName string, username string) {
	_, err := d.db.Exec("UPDATE replays SET username2=? WHERE fullReplayName=?", username, replayName)
	if err != nil {
		panic(err)
	}
}

// User represents the all of the data stored about a user, less their password
type Replay struct {
	GameName       string
	FullReplayName string
	Time           string
	Username1      string
	Username2      string
}

func (d *DB) GetReplay(replay string) (*Replay, error) {
	var createdAt string
	var name string
	var fullReplayName string
	var username1 string
	var username2 string
	err := d.db.QueryRow(`SELECT createdAt, gameName, fullReplayName, username1, username2 FROM replays WHERE fullReplayName=?`,
		replay).Scan(&createdAt, &name, &fullReplayName, &username1, &username2)
	if err != nil {
		return nil, err
	}
	return &Replay{
		Time:           createdAt,
		GameName:       name,
		FullReplayName: fullReplayName,
		Username1:      username1,
		Username2:      username2,
	}, nil
}

// get all replays in reverse chronological

// TODO: games that are not started because only one player entered will be added to the database
// but they will not be displayed on the frontend because the template automatically checks for this fortunately
func (d *DB) GetReplays() ([]Replay, error) {
	replays := []Replay{}
	rows, err := d.db.Query(`SELECT createdAt, gameName, fullReplayName, username1, username2 
FROM replays ORDER BY id DESC LIMIT 0, 50`) // returns the last 50 games
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var createdAt string
		var name string
		var fullReplayName string
		var username1 string
		var username2 string
		err = rows.Scan(&createdAt, &name, &fullReplayName, &username1, &username2)
		if err != nil {
			return nil, err
		}
		replays = append(replays, Replay{
			Time:           createdAt,
			GameName:       name,
			FullReplayName: fullReplayName,
			Username1:      username1,
			Username2:      username2,
		})
	}
	return replays, nil
}
