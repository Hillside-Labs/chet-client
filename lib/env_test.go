package lib

import (
	"log"
	"testing"
)

func TestFindRepo(t *testing.T) {
	env := &CommandEnv{}

	err := FindRepo(env)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(env)

	if env.Repo == "" {
		t.Fatal(env)
	}

	if env.Branch == "" {
		t.Fatal(env)
	}
}
