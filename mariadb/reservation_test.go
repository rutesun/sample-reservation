package mariadb

import (
	"testing"

	"fmt"

	"time"

	"github.com/rutesun/reservation/config"
	"github.com/rutesun/reservation/exception"
	"github.com/stretchr/testify/assert"
)

var mariadb *db

func init() {
	con, err := config.Parse()
	if err != nil {
		panic(err)
	}
	fmt.Printf("\n%+v\n", con)
	setting, err := config.Make(con)
	if err = setting.DB.Ping(); err != nil {
		panic(err)
	}
	mariadb = New(setting.DB)
}

func TestDb_listAll(t *testing.T) {
	list, err := mariadb.listAll(time.Now())

	assert.NoError(t, err)

	for item := range list {
		t.Logf("detail = %+v", item)
	}
}

func TestDb_parseTime(t *testing.T) {
	date, err := time.Parse(time.RFC3339, "2018-08-02T00:00:00+09:00")
	assert.NoError(t, err)

	parsedTime, err := concatHhmm(date, 930)
	assert.NoError(t, err)

	hh := 930 / 100
	mm := 930 % 100
	interval := time.Duration(mm)*time.Minute + time.Duration(hh)*time.Hour

	assert.Equal(t, parsedTime, date.Add(interval))
}

func TestDb_convert(t *testing.T) {
	dto := &dtoReservation{
		ID:         1,
		RoomID:     1,
		RoomName:   "회의실A",
		UserName:   "Ted",
		TargetDate: time.Now(),
		StartTime:  900, EndTime: 1100,
	}

	detail := convertReservation(dto)

	assert.Equal(t, dto.RoomID, detail.Room.ID)
	assert.Equal(t, dto.RoomName, detail.Room.Name)
	assert.Equal(t, dto.UserName, detail.User.Name)
	pt, err := concatHhmm(dto.TargetDate, dto.StartTime)
	assert.NoError(t, err)
	assert.Equal(t, pt, detail.Start)
	pt, err = concatHhmm(dto.TargetDate, dto.EndTime)
	assert.Equal(t, pt, detail.End)

	t.Logf("%+v", detail)
}

func TestDb_CheckAvailable(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T00:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-04T23:00:00+09:00")

	_, err := mariadb.CheckAvailable(1, st, et)
	assert.NoError(t, err)
}

func TestDb_Make(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T18:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-04T19:00:00+09:00")

	err := mariadb.Make(1, 1, st, et)
	if err != nil {
		assert.EqualError(t, err, exception.Unavailable.Error())
	}
}

func TestDb_Cancel(t *testing.T) {

	_, err := mariadb.Cancel(1)

	assert.NoError(t, err)
}
