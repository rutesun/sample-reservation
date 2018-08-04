package mariadb

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/rutesun/reservation/log"
	"gopkg.in/Masterminds/squirrel.v1"
)

type queryFn func(interface{}, string, ...interface{}) error

func (db *db) query(v interface{}, q squirrel.SelectBuilder, fn queryFn) error {
	query, args, err := q.ToSql()
	if err != nil {
		return err
	}

	log.Debugf("query = %s\targs = %v", query, args)

	return fn(v, query, args...)
}

func (db *db) Query(q squirrel.SelectBuilder) (*sqlx.Rows, error) {
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	return db.DB.Queryx(query, args)
}

func (db *db) Get(v interface{}, q squirrel.SelectBuilder) error {
	return db.query(v, q, db.DB.Get)
}

func (db *db) Select(v interface{}, q squirrel.SelectBuilder) error {
	return db.query(v, q, db.DB.Select)
}

type toSql interface {
	ToSql() (string, []interface{}, error)
}

func (db *db) Exec(q toSql) (sql.Result, error) {
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	log.Debugf("query = %s\targs = %v", query, args)

	return db.DB.Exec(query, args...)
}
