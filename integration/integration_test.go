package integration

import (
	"fmt"
	"testing"
	"time"

	"reflect"

	"github.com/rutesun/reservation/config"
	"github.com/rutesun/reservation/exception"
	"github.com/rutesun/reservation/mariadb"
	"github.com/rutesun/reservation/reservation"
	"github.com/stretchr/testify/assert"
)

var service *reservation.Service

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
	mariadb := mariadb.New(setting.DB)
	service = reservation.New(mariadb)
}

var (
	roomID   = int64(1)
	userName = "Ted"

	date, _ = time.Parse(time.RFC3339, "2018-08-07T0:00:00+09:00")
)

func TestReservation_RoomList(t *testing.T) {
	rooms, err := service.RoomList()
	assert.NoError(t, err)

	for _, r := range rooms {
		t.Logf("room: %+v", r)
	}
}

func TestReservation_List(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T18:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-30T18:00:00+09:00")
	_, err := service.List(st, et)
	assert.NoError(t, err)
}

func TestReservation_Available(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-05T16:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-05T00:00:00+09:00")

	_, err := service.Available(1, st, et)
	assert.EqualError(t, err, exception.InvalidRequest.Error())
}

func TestReservation_Make(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-07T16:00:00+09:00")

	t.Run("Invalid Request: 끝나는 시간이 시작 시간 보다 앞설 때 ", func(t *testing.T) {
		et, _ := time.Parse(time.RFC3339, "2018-08-07T00:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		assert.EqualError(t, err, exception.InvalidRequest.Error())

	})

	t.Run("Invalid Request: 정시, 30분 단위가 아닐 때", func(t *testing.T) {
		et, _ := time.Parse(time.RFC3339, "2018-08-08T16:10:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		assert.EqualError(t, err, exception.InvalidRequest.Error())

	})

	t.Run("정상 예약", func(t *testing.T) {
		et, _ := time.Parse(time.RFC3339, "2018-08-07T19:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		if err != nil {
			assert.EqualError(t, err, exception.Unavailable.Error())
		}
	})

	t.Run("끝나는 시간과 시작 시간이 겹칠 때", func(t *testing.T) {
		st, _ := time.Parse(time.RFC3339, "2018-08-07T14:00:00+09:00")
		et, _ := time.Parse(time.RFC3339, "2018-08-07T16:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		if err != nil {
			assert.EqualError(t, err, exception.Unavailable.Error())
		}
	})

	t.Run("반복 예약", func(t *testing.T) {
		st, _ := time.Parse(time.RFC3339, "2018-08-07T12:00:00+09:00")
		et, _ := time.Parse(time.RFC3339, "2018-08-07T14:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{Repeat: 10})
		if err != nil {
			assert.EqualError(t, err, exception.Unavailable.Error())
		}
	})

}

func TestReservation_Integration(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-07T10:00:00+09:00")
	et, _ := time.Parse(time.RFC3339, "2018-08-07T12:00:00+09:00")

	err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
	if err != nil {
		assert.EqualError(t, err, exception.Unavailable.Error())
	}

	st, _ = time.Parse(time.RFC3339, "2018-08-07T11:00:00+09:00")
	et, _ = time.Parse(time.RFC3339, "2018-08-07T12:00:00+09:00")

	err = service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
	assert.EqualError(t, err, exception.Unavailable.Error())

	reservedMap, err := service.List(st, st.AddDate(0, 0, 1))
	assert.NoError(t, err)

	list, ok := reservedMap[roomID]
	assert.True(t, ok)
	assert.True(t, len(list) > 0)

	detail := list[0]
	assert.True(t, detail.ID > 0)
	assert.True(t, detail.Room.ID > 0)

	for _, detail := range list {
		_, err = service.Cancel(detail.ID)
		assert.NoError(t, err)
	}

	reservedMap, err = service.List(st, st.AddDate(0, 0, 1))
	keys := reflect.ValueOf(reservedMap).MapKeys()
	assert.Equal(t, len(keys), 0)
}
