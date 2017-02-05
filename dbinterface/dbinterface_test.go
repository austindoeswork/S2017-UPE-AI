package dbinterface

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestUpdateELO(t *testing.T) {
	os.Chdir("..")
	db := NewDB()
	defer db.Close()
	const username1 = "poe"
	const username2 = "bibbles"
	newELO1, newELO2, err := db.UpdateELO(username1, username2, 1)
	if err != nil {
		log.Fatalln(err)
		t.FailNow()
	}
	fmt.Printf("%s's new ELO: %f\n", username1, *newELO1)
	fmt.Printf("%s's new ELO: %f\n", username2, *newELO2)
}
