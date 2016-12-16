package server

import (
	"bufio"
	"os"
	"math/rand"
	"log"
	"io/ioutil"
	// "strconv"
	// "https://github.com/gtank/bloomfilter/blob/master/bloomfilter"
)

func toUpper(x byte) byte {
	if x >= 'a' && x <= 'z' {
		return x - 'a' + 'A'
	} else {
		return x
	}
}

// TODO add uint count as arg to get multiple bools
func getWordsFromFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	// first, use counter to find length of file
	// TODO split from this function so in the case of duplicate keys, this isn't called twice
	counter := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		scanner.Text()
		counter++
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// log.Println("counter: " + strconv.Itoa(counter))
	// use random to find a line to grab
	selected := rand.Intn(counter)
	file.Seek(0, 0)
	counter = 0
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if selected == counter { // is this the line? return
			bytetext := append([]byte{toUpper(text[0])}, []byte(text[1:])...)
			return bytetext
		}
		counter++
	}
	// log.Println("read file " + string(path[:]) + ", length = " + string(counter))
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return nil // we should never be reaching here, otherwise counter is outside file range
}

// TODO change adj to uint to allow for grabbing multiple adjectives from file without
// opening/closing the files unnecessarily
// note the difference between generating A key and generating a UNIQUE key (different functions)
func generateKey(name bool, adverbs bool, adj bool, noun bool) []byte {
	var key []byte
	if (name) {
		key = append(key, append(getWordsFromFile("server/words/names.txt"), byte('s'))...)
	}
	if (adverbs) {
		key = append(key, getWordsFromFile("server/words/adverbs.txt")...)
	}
	if (adj) {
		key = append(key, getWordsFromFile("server/words/adjs.txt")...)
	}
	if (noun) {
		key = append(key, getWordsFromFile("server/words/nouns.txt")...)
	}
	// log.Println("current key: " + string(key[:]))
	return key
}

// TODO do i even need to check for file existence with OS.create()?
// returns TRUE if key is unique, will also append if so (if file does not exist, it will initialize it as well)
// returns FALSE if key is not unique, key has obviously not been added in this case
func addKey(key []byte) bool {
	path := "server/words/used/" + string(key[0]) + ".txt"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// file does not exist, init
		err = ioutil.WriteFile(path, append(key, []byte{'\\','n'}...), 0777)
		if err != nil {
			log.Fatal(err)
		}
		return true
	} else {
		// file exists, check it
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			text := scanner.Text()
			if text == string(key) {
				return false
			}
		}
		// did not find in corresponding file, add and return true
		file, err = os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		
		file.Write(append(key, []byte("\n")...)) // TODO currently overrides entire file :(
		file.Sync()
		return true
	}
	return false // shouldn't really be returning here except when it panics?
}

// main generate key function, guarantees uniqueness
func GenerateUniqueKey(name bool, adverbs bool, adj bool, noun bool) []byte {
	key := generateKey(name, adverbs, adj, noun)
	for !addKey(key) {
		key = generateKey(name, adverbs, adj, noun)
	}
	return key
}
