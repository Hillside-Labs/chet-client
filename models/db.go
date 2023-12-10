package models

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

type DSN struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
	SSLMode  string
}

func Connect(fn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(fn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Record{}, &LocalConfig{})

	return db, err
}

func (dsn *DSN) String() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dsn.Host, dsn.User, dsn.Password, dsn.DBName, dsn.Port, dsn.SSLMode,
	)
}

func NewDSN(dburi string) (*DSN, error) {
	url, err := url.Parse(dburi)
	if err != nil {
		return nil, err
	}

	pw, _ := url.User.Password()
	port := 5432
	if url.Port() != "" {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			return nil, err
		}
	}

	path := strings.Split(strings.TrimLeft(url.Path, "/"), "/")

	if len(path) == 0 {
		return nil, fmt.Errorf("missing db name: %s", url.Path)
	}

	dbname := path[0]

	sslmode := url.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "disable"
	}

	return &DSN{
		Host:     url.Hostname(),
		User:     url.User.Username(),
		Password: pw,
		DBName:   dbname,
		Port:     port,
		SSLMode:  sslmode,
	}, nil
}
