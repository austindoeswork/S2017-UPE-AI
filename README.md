# UPE AI COMP SPRING 2017

## Installation

Really convenient way to install Go: http://www.hostingadvice.com/how-to/install-golang-on-ubuntu/ (-Darwin)

Note that this server uses MySQL as its database, make sure mysql-server is installed.	
Additionally, you may need to edit dbinterface/CREDENTIALS to have proper credentials. (i.e. replace username, password as necessary). If the server does not detect a CREDENTIALS file, it will use the default, which is "root" user with no password.

(TODO: add dbinterface/CREDENTIALS to .gitignore?)

The server will automatically use that account to create a database called aicomp if it doesn't exist, and a table called users within if that doesn't exist either.
It will not override existing databases and tables of those names. You may need to drop the database manually if changes occur in the schemas?

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

/dbinterface = Acts as a wrapper around the MySQL driver

/game = Game objects are the representations of the internals of games (currently has pong and tdef).

/gamemanager = Contains game wrappers and managers. The former is an object that multiplexes game object output to all available listeners.
People who want to spectate/play ongoing matches will interface through the game wrappers. Game managers hold all of the wrapper/game pairings.

/server = Server details (routing, etc). Also holds keygen.go, which generates the api keys for players upon signup.

/templates = These .html files are loaded by the template manager on server startup. (Coming soon: template reloader when these templates change, so that server doesn't need to reload on each startup)
Note that the header.html and footer.html templates that are included with each of these templates start and end the main container div that are pretty much used everywhere.

## TODOLIST

Victory screen + stats post-game

Status splash screen before game

Status message over websockets pre-game

Death anims (fade?)