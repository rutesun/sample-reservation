package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	host     = "ted.ck5mrdxgowlk.ap-northeast-2.rds.amazonaws.com"
	password = "ePix9L5ILrw3"
)

func TestDefaultParse(t *testing.T) {
	con, err := Parse()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\nConfig = %+v\n", con)
	assert.Equal(t, con.Host, "0.0.0.0")
	assert.Equal(t, con.Port, 8080)
	assert.Equal(t, con.Database.User, "root")
	assert.Equal(t, con.Database.Port, 3306)
	assert.Empty(t, con.Database.Password)
}

func TestParse(t *testing.T) {
	err := os.Setenv("DATABASE_HOST", host)
	if err != nil {
		t.Fatal(err)
	}

	con, err := Parse()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, con.Database.Host, host)
}

func TestMake(t *testing.T) {
	var err error
	err = os.Setenv("DATABASE_HOST", host)
	err = os.Setenv("DATABASE_PASSWORD", password)
	err = os.Setenv("DATABASE_NAME", "reservation")

	if err != nil {
		t.Fatal(err)
	}

	con, err := Parse()
	if err != nil {
		t.Fatal(err)
	}

	set, err := Make(con)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("\nSetting = %+v", set)
	assert.NotNil(t, set.DB)
}
