# UPE AI COMP SPRING 2017

## Installation

- Get the code

```
git clone https://github.com/austindoeswork/S2017-UPE-AI.git
```

- Run the code (if you don't have golang, go <a href="https://golang.org/doc/install/source">here</a>)

```
go run main.go
```

## FOLDER SPECIFICS

/game = Game objects are the representations of the internals of games (currently pong).

/gamemanager = Contains game wrappers and managers. The former is an object that multiplexes game object output to all available listeners.
People who want to spectate/play ongoing matches will interface through the game wrappers. Game managers hold all of the wrapper/game pairings.

/server = Server details (routing, etc). Also holds keygen.go, which generates the api keys for players upon signup.
Currently overly coupled with the actual MySQL database, this will be decoupled into the database interface soon.

## ROUTES

/signup = create new account given username, password, will give you an apikey

/login = login using existing username, password details, will show you apikey