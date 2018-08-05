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

type customTime time.Time

func (ct customTime) getDateStr() string {
	y, m, d := time.Time(ct).Date()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
}

func (ct customTime) getHhmmInt() int {
	return time.Time(ct).Hour()*100 + time.Time(ct).Minute()
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
		"ru.name AS user_name",
		"r.target_date",
		"r.start_time",
		"r.end_time",
	).
		From("reservation AS r").
		Join("reservation_item AS ri ON r.item_id = ri.id").
		Join("reservation_user AS ru ON r.user_id = ru.id").
		Where("r.target_date = ?", customTime(date).getDateStr())

	err := db.Select(&reservations, builder)
	return reservations, err
}

func (db *db) CheckAvailable(roomID int, startTime time.Time, endTime time.Time) (bool, error) {
	builder := sq.Select("count(*)").
		From("reservation").
		Where("item_id = ?", roomID).
		Where("target_date = ?", customTime(startTime).getDateStr()).
		Where("end_time >= ? AND start_time <= ?", customTime(startTime).getHhmmInt(), customTime(endTime).getHhmmInt())

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

type repeater struct {
	count, max int
}

var emptyRepeater = repeater{}

func (db *db) MakeRepeatly(roomID int, userID int, repeatCnt int, startTime time.Time, endTime time.Time) error {
	var (
		err error
		tx  *sql.Tx
	)

	if tx, err = db.DB.Begin(); err != nil {
		return errors.Wrap(err, "Fail to begin transaction")
	}
	for i := 0; i < repeatCnt; i++ {
		if err := db.make(tx, roomID, userID, startTime, endTime, repeater{i, repeatCnt}); err != nil {
			tx.Rollback()
			return errors.Wrap(err, "예약 반복 생성 중에 실패하였습니다:")
		}
		startTime = startTime.AddDate(0, 0, 7)
		endTime = endTime.AddDate(0, 0, 7)
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "Fail to commit transaction")
	}
	return nil
}

func (db *db) Make(roomID int, userID int, startTime time.Time, endTime time.Time) error {
	if err := db.make(db.DB, roomID, userID, startTime, endTime, emptyRepeater); err != nil {
		return errors.WithStack(err)
	} else {
		return nil
	}
}

func (db *db) make(execer execer, roomID, userID int, startTime, endTime time.Time, repeatable repeater) error {
	var err error

	columes := []string{"item_id", "user_id", "target_date", "start_time", "end_time"}
	values := []interface{}{roomID, userID, customTime(startTime).getDateStr(), customTime(startTime).getHhmmInt(), customTime(endTime).getHhmmInt()}

	if repeatable != emptyRepeater {
		columes = append(columes, "repeat_cnt", "repeat_max")
		values = append(values, repeatable.count, repeatable.max)
	}

	builder := sq.Insert("reservation").
		Columns(columes...).
		Values(values...)

	query, args, err := builder.ToSql()

	log.Debugf("query = %s\targs = %v", query, args)

	if err != nil {
		return errors.WithStack(err)
	}

	if able, err := db.CheckAvailable(roomID, startTime, endTime); err != nil {
		return errors.WithStack(err)
	} else {
		if !able {
			return exception.Unavailable
		}
	}
	if _, err = execer.Exec(query, args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (db *db) Cancel(reservationID uint) (bool, error) {
	builder := sq.Delete("reservation").Where("id = ?", reservationID)
	if _, err := db.Exec(builder); err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}

type dtoUser struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type dtoRoom struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type dtoReservation struct {
	ID         int64     `db:"id"`
	RoomID     int64     `db:"room_id"`
	RoomName   string    `db:"room_name"`
	UserName   string    `db:"user_name"`
	TargetDate time.Time `db:"target_date"`
	StartTime  int       `db:"start_time"`
	EndTime    int       `db:"end_time"`
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
	}
}
