package config

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Host     string `default:"0.0.0.0"`
	Port     int    `default:"8080"`
	Database struct {
		User         string `default:"root"`
		Password     string
		Host         string `default:"0.0.0.0"`
		Port         int    `default:"3306"`
		Name         string
		Charset      string `default:"utf8mb4"`
		Location     string `default:"UTC"`
		MaxIdleConns int    `default:"1"`
		MaxOpenConns int    `default:"10"`
	}
}

func Parse() (*config, error) {
	c := config{}
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}

type setting struct {
	*sqlx.DB
}

func Make(c *config) (*setting, error) {
	endpoint := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		c.Database.User, c.Database.Password,
		c.Database.Host, c.Database.Port, c.Database.Name,
		c.Database.Charset, c.Database.Location)
	db, err := sqlx.Open("mysql", endpoint)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}

	db.SetMaxIdleConns(c.Database.MaxIdleConns)
	db.SetMaxOpenConns(c.Database.MaxOpenConns)

	err = db.Ping()

	return &setting{db}, err
}
