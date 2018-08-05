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

var (
	roomID   = int64(1)
	userName = "Ted"
)

func TestDb_listAll(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T00:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-30T00:00:00+09:00")
	list, err := mariadb.List(st, et)

	assert.NoError(t, err)

	for item := range list {
		t.Logf("detail = %+v", item)
	}
}

func TestDb_convert(t *testing.T) {
	now := time.Now()
	dto := &dtoReservation{
		ID:        1,
		RoomID:    roomID,
		RoomName:  "회의실A",
		UserName:  userName,
		StartTime: now, EndTime: now,
	}

	detail := convertReservation(dto)

	assert.Equal(t, dto.RoomID, detail.Room.ID)
	assert.Equal(t, dto.RoomName, detail.Room.Name)
	assert.Equal(t, dto.StartTime, detail.Start)
	assert.Equal(t, dto.EndTime, detail.End)

	t.Logf("%+v", detail)
}

func TestDb_Make(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T18:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-04T19:00:00+09:00")

	id, err := mariadb.Make(roomID, userName, st, et, "")
	if err != nil {
		assert.EqualError(t, err, exception.Unavailable.Error())
	}

	t.Log(id)

	check, err := mariadb.Available(roomID, st, et)
	assert.NoError(t, err)
	assert.False(t, check)
}

func TestDb_CheckAvailable(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T00:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-04T23:00:00+09:00")

	_, err := mariadb.Available(roomID, st, et)
	assert.NoError(t, err)
}

func TestDb_MakeRepeatly(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-05T16:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-05T19:00:00+09:00")

	repeatCnt := 5
	ids, err := mariadb.MakeRepeatly(roomID, userName, st, et, repeatCnt, "")
	if err != nil {
		assert.EqualError(t, err, exception.Unavailable.Error())
	}

	t.Log(ids)

	for i := 0; i < 5; i++ {
		check, err := mariadb.Available(roomID, st, et)
		assert.NoError(t, err)
		assert.False(t, check)
		st = st.AddDate(0, 0, 7)
		et = et.AddDate(0, 0, 7)
	}
}

func TestDb_Cancel(t *testing.T) {

	_, err := mariadb.Cancel(1)

	assert.NoError(t, err)
}
