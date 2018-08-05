package reservation

import (
	"time"

	"github.com/pkg/errors"
	"github.com/rutesun/reservation/exception"
	"github.com/rutesun/reservation/log"
)

const DateFormat = "2016-01-02"

type Detail struct {
	ID    int64     `json:"id"`
	Room  Room      `json:"room"`
	User  string    `json:"user"`
	Start time.Time `json:"startTime"`
	End   time.Time `json:"endTime"`
	Memo  string    `json:"memo"`
}

type ExtraInfo struct {
	Memo   string
	Repeat int
}

var emptyExtra = ExtraInfo{}

type Reservations []*Detail

type ReservedMap map[int64][]*Detail

func NewReservedMap() *ReservedMap {
	rMap := make(ReservedMap)
	return &rMap
}

type reservationRepository interface {
	RoomList() ([]*Room, error)
	List(startDate, endDate time.Time) ([]*Detail, error)
	Available(roomID int64, startTime, endTime time.Time) (bool, error)
	Make(roomID int64, userName string, startTime, endTime time.Time, memo string) (int64, error)
	MakeRepeatly(roomID int64, userName string, startTime, endTime time.Time, repeatCnt int, memo string) ([]int64, error)
	Cancel(reservationID int64) (bool, error)
}

type Service struct {
	reservation reservationRepository
}

func New(reservation reservationRepository) *Service {
	return &Service{reservation}
}

func (s *Service) RoomList() ([]*Room, error) {
	return s.reservation.RoomList()
}

func (s *Service) List(startDate, endDate time.Time) (map[int64][]*Detail, error) {
	if endDate.Before(startDate) {
		return nil, errors.WithStack(exception.InvalidRequest)
	}

	list, err := s.reservation.List(startDate, endDate)
	if err != nil {
		return nil, err
	}

	log.Debugf("From: %v - To: %v, 예약 Count: %d", startDate, endDate, len(list))
	reservedMap := make(map[int64][]*Detail)

	for _, detail := range list {
		if l, ok := reservedMap[detail.Room.ID]; !ok {
			l = []*Detail{detail}
			reservedMap[detail.Room.ID] = l
		} else {
			reservedMap[detail.Room.ID] = append(l, detail)
		}
	}
	return reservedMap, nil
}

func (s *Service) Available(roomID int64, startTimestamp time.Time, endTimestamp time.Time) (bool, error) {
	if endTimestamp.Before(startTimestamp) {
		return false, errors.WithStack(exception.InvalidRequest)
	}

	return s.reservation.Available(roomID, startTimestamp, endTimestamp)
}

func (s *Service) Make(roomID int64, userName string, startTimestamp time.Time, endTimestamp time.Time, extra ExtraInfo) error {
	if endTimestamp.Before(startTimestamp) {
		return errors.WithStack(exception.InvalidRequest)
	}

	// 정시 or 30분 단위로만 예약 가능
	if startTimestamp.Minute()%30 != 0 ||
		endTimestamp.Minute()%30 != 0 {
		return errors.WithStack(exception.InvalidRequest)
	}
	var err error
	if extra.Repeat > 1 {
		_, err = s.reservation.MakeRepeatly(roomID, userName, startTimestamp, endTimestamp, extra.Repeat, extra.Memo)
	} else {
		_, err = s.reservation.Make(roomID, userName, startTimestamp, endTimestamp, extra.Memo)
	}
	return errors.WithStack(err)

}

func (s *Service) Cancel(reservationID int64) (bool, error) {
	return s.reservation.Cancel(reservationID)
}
