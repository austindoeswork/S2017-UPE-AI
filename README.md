# UPE AI COMP SPRING 2017

## Installation

Note that this server uses MySQL as its database, make sure mysql-server is installed.
Additionally, the current implementation requires that there exists an aicomp database with a single table called "users". The following command was used to create the database. On my machine I use the username root with no password, although you will need to change the internals of main.go if your MySQL creds are different.

(Coming soon: bash script that will create new DB, or server automatically creates MySQL db on startup)

```
CREATE TABLE users(
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50),
    password VARCHAR(120),
    apikey VARCHAR(50)
);
```

- Get the code

```
git clone https://github.com/austindoeswork/S2017-UPE-AI.git
```

- Run the code (if you don't have golang, go <a href="https://golang.org/doc/install/source">here</a>)

```
go run main.go
```

- To build a binary, run in this order:

```
go get
go build
./S2017-UPE-AI
```

## FOLDER SPECIFICS

/game = Game objects are the representations of the internals of games (currently pong).

/gamemanager = Contains game wrappers and managers. The former is an object that multiplexes game object output to all available listeners.
People who want to spectate/play ongoing matches will interface through the game wrappers. Game managers hold all of the wrapper/game pairings.

/server = Server details (routing, etc). Also holds keygen.go, which generates the api keys for players upon signup.
Currently overly coupled with the actual MySQL database, this will be decoupled into the database interface soon.

## ROUTES

/ = index, you can play matches of tower defense (press up to spawn unit, colors correspond to %hp of unit)

/signup = create new account given username, password, will give you an apikey

/login = login using existing username, password details, will redirect to /profile

/logout = logs account out

/profile = shows you apikey as long as you have a valid login cookie
