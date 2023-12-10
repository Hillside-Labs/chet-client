package models

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	Label     string
	Cmd       string
	Duration  time.Duration
	Repo      string
	Branch    string
	Username  string
	OS        string
	Container bool
}

func (r Record) String() {
	fmt.Printf("Record:%+v\n", r)
}

type LocalConfig struct {
	gorm.Model
	UserEmail     string
	TeamName      string
	ClientID      string `gorm:"unique"`
	ClientSecret  string
	ServerAddress string
	Token         string
	DisableRemote bool
}

func (lc LocalConfig) String() {
	fmt.Printf("LocalConfig:%+v\n", lc)
}

func Connect(fn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(fn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Record{}, &LocalConfig{})

	return db, err
}
