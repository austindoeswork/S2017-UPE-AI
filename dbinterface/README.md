This folder contains the DB Interface, which acts as a wrapper around the Golang MySQL Driver.
We've added the CREDENTIALS file to the .gitignore, but it is required for the server to run (or at least, unless you're okay with the default user (root, no pass))

CREDENTIALS FILE SHOULD LOOK LIKE:
user:<username>
pass:<password>