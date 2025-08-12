package main_test

import (
	"bot/sql"
	"log"
	"testing"
)

func TestFunctions(t *testing.T) {
	prefix, err := sql.GetPrefix()
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Current Prefix is:", prefix)

}
