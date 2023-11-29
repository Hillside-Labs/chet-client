package models

import (
	"fmt"
	"strings"
	"testing"
)

func TestDSNWithSSLEnabled(t *testing.T) {
	dburi := "postgresql://user:pw@pgpool.xample.com/chetapp?sslmode=require"
	dsn, err := NewDSN(dburi)
	if err != nil {
		t.Fatal(err)
	}

	fields := strings.Fields(dsn.String())

	for _, field := range fields {
		f := strings.SplitN(field, "=", 2)
		k := f[0]
		v := f[1]

		switch k {
		case "host":
			if v != dsn.Host {
				t.Errorf("dsn host: expected %s got %s", dsn.Host, v)
			}
		case "user":
			if v != dsn.User {
				t.Errorf("dsn user: expected %s got %s", dsn.User, v)
			}
		case "password":
			if v != dsn.Password {
				t.Errorf("dsn password: expected %s got %s", dsn.Password, v)
			}
		case "dbname":
			if v != dsn.DBName {
				t.Errorf("dsn dbname: expected %s got %s", dsn.DBName, v)
			}
		case "port":
			if v != fmt.Sprintf("%d", dsn.Port) {
				t.Errorf("dsn port: expected %d got %s", dsn.Port, v)
			}
		case "sslmode":
			if v != dsn.SSLMode {
				t.Errorf("dsn sslmode: expected %s got %s", dsn.SSLMode, v)
			}
		}

	}
}
