package mariadb

import (
	"time"

	"fmt"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rutesun/reservation/exception"
	"github.com/rutesun/reservation/log"
	"github.com/rutesun/reservation/reservation"
	sq "gopkg.in/Masterminds/squirrel.v1"
)

type db struct {
	DB *sqlx.DB
}

func New(d *sqlx.DB) *db {
	return &db{d}
}

func (db *db) ListAll(date time.Time) ([]*reservation.Detail, error) {
	list, err := db.listAll(date)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	details := make([]*reservation.Detail, len(list))

	for i, o := range list {
		details[i] = convertReservation(o)
	}
	return details, nil
}

func (db *db) listAll(date time.Time) ([]*dtoReservation, error) {
	reservations := []*dtoReservation{}
	builder := sq.Select(
		"r.id",
		"ri.id AS room_id",
		"ri.name AS room_name",
		"r.user_name AS user_name",
		"r.target_date",
		"r.start_time",
		"r.end_time",
		"r.memo",
	).
		From("reservation AS r").
		Join("reservation_item AS ri ON r.item_id = ri.id").
		Where("r.target_date = ?", date)

	err := db.Select(&reservations, builder)
	return reservations, err
}

func (db *db) Available(roomID int64, date time.Time, startTime, endTime int) (bool, error) {
	builder := sq.Select("count(*)").
		From("reservation").
		Where("item_id = ?", roomID).
		Where("target_date = ?", reservation.CustomTime(date).GetDateStr()).
		Where("end_time >= ? AND start_time <= ?", startTime, endTime)

	count := 0
	if err := db.Get(&count, builder); err != nil {
		return false, errors.WithStack(err)
	}

	if count == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

type execer interface {
	Exec(string, ...interface{}) (sql.Result, error)
}

func (db *db) MakeRepeatly(roomID int64, userName string, date time.Time, startTime, endTime int, repeatCnt int, memo string) ([]int64, error) {
	var (
		err error
		tx  *sql.Tx
	)

	ids := []int64{}
	if tx, err = db.DB.Begin(); err != nil {
		return nil, errors.Wrap(err, "Fail to begin transaction")
	}
	if repeatCnt == 0 {
		return nil, exception.InvalidRequest
	}
	for i := 0; i < repeatCnt; i++ {
		if res, err := db.make(tx, roomID, userName, date, startTime, endTime,
			fmt.Sprintf("(반복 %d/%d회)\n%s", i+1, repeatCnt, memo)); err != nil {
			tx.Rollback()
			return nil, err
		} else {
			id, _ := res.LastInsertId()
			ids = append(ids, id)
			date = date.AddDate(0, 0, 7)
		}

	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "Fail to commit transaction")
	}
	return ids, nil
}

func (db *db) Make(roomID int64, userName string, date time.Time, startTime, endTime int, memo string) (int64, error) {
	if res, err := db.make(db.DB, roomID, userName, date, startTime, endTime, memo); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return res.LastInsertId()
	}
}

func (db *db) make(execer execer, roomID int64, userName string, date time.Time, startTime, endTime int, memo string) (sql.Result, error) {
	var err error

	columns := []string{"item_id", "user_name", "target_date", "start_time", "end_time", "memo"}
	values := []interface{}{roomID, userName, date, startTime, endTime, memo}

	builder := sq.Insert("reservation").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.ToSql()

	log.Debugf("query = %s\targs = %v", query, args)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if able, err := db.Available(roomID, date, startTime, endTime); err != nil {
		return nil, errors.WithStack(err)
	} else {
		if !able {
			return nil, exception.Unavailable
		}
	}
	return execer.Exec(query, args...)
}

func (db *db) Cancel(reservationID int64) (bool, error) {
	builder := sq.Delete("reservation").Where("id = ?", reservationID)
	if _, err := db.Exec(builder); err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}

type dtoRoom struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type dtoReservation struct {
	ID         int64          `db:"id"`
	RoomID     int64          `db:"room_id"`
	RoomName   string         `db:"room_name"`
	UserName   string         `db:"user_name"`
	TargetDate time.Time      `db:"target_date"`
	StartTime  int            `db:"start_time"`
	EndTime    int            `db:"end_time"`
	Memo       sql.NullString `db:"memo"`
}

// RFC3339 format
const timeFormat = "%d-%02d-%02dT%02d:%02d:00+09:00"

func concatHhmm(date time.Time, hhmmNumber int) (time.Time, error) {
	mm := hhmmNumber % 100
	hh := hhmmNumber / 100
	y, m, d := date.Date()
	return time.Parse(time.RFC3339,
		fmt.Sprintf(timeFormat, y, m, d, hh, mm))
}

func convertReservation(r *dtoReservation) *reservation.Detail {
	if r == nil {
		return nil
	}
	startTime, _ := concatHhmm(r.TargetDate, r.StartTime)
	endTime, _ := concatHhmm(r.TargetDate, r.EndTime)

	return &reservation.Detail{
		Room: reservation.Room{
			ID:   r.RoomID,
			Name: r.RoomName,
		},
		User: reservation.User{
			Name: r.UserName,
		},
		Start: startTime, End: endTime,
		Memo: r.Memo.String,
	}
}
