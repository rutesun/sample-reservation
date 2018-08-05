package integration

import (
	"fmt"
	"testing"
	"time"

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
)

func TestReservation_List(t *testing.T) {
	st, _ := time.Parse(time.RFC3339, "2018-08-04T18:00:00+09:00")
	_, err := service.ListAll(st)
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

	t.Run("Invalid Request: 시작 날짜와 끝나는 날짜가 다를 때", func(t *testing.T) {
		et, _ := time.Parse(time.RFC3339, "2018-08-08T19:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		assert.EqualError(t, err, exception.InvalidRequest.Error())

	})

	t.Run("정상 동작", func(t *testing.T) {
		et, _ := time.Parse(time.RFC3339, "2018-08-07T19:00:00+09:00")

		err := service.Make(roomID, userName, st, et, reservation.ExtraInfo{})
		if err != nil {
			assert.EqualError(t, err, exception.Unavailable.Error())
		}

	})
}

func TestReservation_Integration(t *testing.T) {

}
