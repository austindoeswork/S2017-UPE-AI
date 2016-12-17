This folder represents all of the currently used keys in the system.

The actual key-list is in the gitignore, this README exists to make sure that git acknowledges that this folder should be a thing.

When keygen.go generates a new key, it bucket sorts it into *.txt, where * is the first letter of the generated key. Of course, if that
key already exists in its file then it will generate another one.