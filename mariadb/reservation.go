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
	return &db{DB: d}
}

func (db *db) RoomList() ([]*reservation.Room, error) {
	rList := []*dtoRoom{}

	builder := sq.Select(
		"r.id",
		"r.name",
	).
		From("reservation_item AS r").Where("r.item_type = 'MEETING'")

	err := db.Select(&rList, builder)

	rooms := make([]*reservation.Room, len(rList))
	for i, r := range rList {
		rooms[i] = convertRoom(r)
	}

	return rooms, err
}

func (db *db) List(startDate, endDate time.Time) ([]*reservation.Detail, error) {
	list, err := db.listAll(startDate, endDate)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	details := make([]*reservation.Detail, len(list))

	for i, o := range list {
		details[i] = convertReservation(o)
	}
	return details, nil
}

func (db *db) listAll(startDate, endDate time.Time) ([]*dtoReservation, error) {
	reservations := []*dtoReservation{}

	builder := sq.Select(
		"r.id",
		"ri.id AS room_id",
		"ri.name AS room_name",
		"r.user_name AS user_name",
		"r.start_time",
		"r.end_time",
		"r.memo",
	).
		From("reservation AS r").
		Join("reservation_item AS ri ON r.item_id = ri.id").
		Where("r.start_time >= ? AND r.end_time < ?", startDate, endDate)

	err := db.Select(&reservations, builder)
	return reservations, err
}

func (db *db) Available(roomID int64, startTime, endTime time.Time) (bool, error) {
	builder := sq.Select("count(*)").
		From("reservation").
		Where("item_id = ?", roomID).
		Where("end_time >= ? AND start_time < ?", startTime, endTime)

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

func (db *db) MakeRepeatly(roomID int64, userName string, startTime, endTime time.Time, repeatCnt int, memo string) ([]int64, error) {
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
		if res, err := db.make(tx, roomID, userName, startTime, endTime,
			fmt.Sprintf("(반복 %d/%d회)\n%s", i+1, repeatCnt, memo)); err != nil {
			tx.Rollback()
			return nil, err
		} else {
			id, _ := res.LastInsertId()
			ids = append(ids, id)
			startTime = startTime.AddDate(0, 0, 7)
			endTime = endTime.AddDate(0, 0, 7)
		}

	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "Fail to commit transaction")
	}
	return ids, nil
}

func (db *db) Make(roomID int64, userName string, startTime, endTime time.Time, memo string) (int64, error) {
	if res, err := db.make(db.DB, roomID, userName, startTime, endTime, memo); err != nil {
		return 0, errors.WithStack(err)
	} else {
		return res.LastInsertId()
	}
}

func (db *db) make(execer execer, roomID int64, userName string, startTime, endTime time.Time, memo string) (sql.Result, error) {
	var err error

	columns := []string{"item_id", "user_name", "start_time", "end_time", "memo"}
	values := []interface{}{roomID, userName, startTime, endTime, memo}

	builder := sq.Insert("reservation").
		Columns(columns...).
		Values(values...)

	query, args, err := builder.ToSql()

	log.Debugf("query = %s\targs = %v", query, args)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if able, err := db.Available(roomID, startTime, endTime); err != nil {
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
	ID        int64          `db:"id"`
	RoomID    int64          `db:"room_id"`
	RoomName  string         `db:"room_name"`
	UserName  string         `db:"user_name"`
	StartTime time.Time      `db:"start_time"`
	EndTime   time.Time      `db:"end_time"`
	Memo      sql.NullString `db:"memo"`
}

func convertRoom(r *dtoRoom) *reservation.Room {
	return &reservation.Room{
		r.ID, r.Name,
	}
}

func convertReservation(r *dtoReservation) *reservation.Detail {
	if r == nil {
		return nil
	}

	return &reservation.Detail{
		ID: r.ID,
		Room: reservation.Room{
			ID:   r.RoomID,
			Name: r.RoomName,
		},
		User:  r.UserName,
		Start: r.StartTime, End: r.EndTime,
		Memo: r.Memo.String,
	}
}
