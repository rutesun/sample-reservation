package reservation

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rutesun/reservation/exception"
)

type Detail struct {
	Room  Room      `json:"room"`
	User  User      `json:"user"`
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

type Repository interface {
	ListAll(date time.Time) ([]*Detail, error)
	Available(roomID int64, date time.Time, startTime, endTime int) (bool, error)
	Make(roomID int64, userName string, date time.Time, startTime, endTime int, memo string) (int64, error)
	MakeRepeatly(roomID int64, userName string, date time.Time, startTime, endTime int, repeatCnt int, memo string) ([]int64, error)
	Cancel(reservationID int64) (bool, error)
}

type Service struct {
	Repository
}

func New(r Repository) *Service {
	return &Service{r}
}

func (s *Service) List(date time.Time) (ReservedMap, error) {
	list, err := s.Repository.ListAll(date)
	if err != nil {
		return nil, err
	}
	reservedMap := make(map[int64][]*Detail)

	for _, detail := range list {
		if l, ok := reservedMap[detail.Room.ID]; !ok {
			reservedMap[detail.Room.ID] = []*Detail{}
			l = append(l, detail)
		} else {
			l = append(l, detail)
		}
	}
	return reservedMap, nil
}

func (s *Service) Available(roomID int64, startTimestamp time.Time, endTimestamp time.Time) (bool, error) {
	if endTimestamp.Before(startTimestamp) {
		return false, errors.WithStack(exception.InvalidRequest)
	}

	var (
		startTime = CustomTime(startTimestamp).GetHhmmInt()
		endTime   = CustomTime(endTimestamp).GetHhmmInt()
	)

	return s.Repository.Available(roomID, startTimestamp, startTime, endTime)
}

func (s *Service) Make(roomID int64, userName string, startTimestamp time.Time, endTimestamp time.Time, extra ExtraInfo) error {
	if endTimestamp.Before(startTimestamp) {
		return errors.WithStack(exception.InvalidRequest)
	}

	var (
		startDateStr = CustomTime(startTimestamp).GetDateStr()
		endDateStr   = CustomTime(endTimestamp).GetDateStr()
		startTime    = CustomTime(startTimestamp).GetHhmmInt()
		endTime      = CustomTime(endTimestamp).GetHhmmInt()
	)

	if startDateStr != endDateStr {
		return errors.WithStack(exception.InvalidRequest)
	}

	// 정시 or 30분 단위로만 예약 가능
	if startTime%100%30 != 0 ||
		endTime%100%30 != 0 {
		return errors.WithStack(exception.InvalidRequest)
	}
	var err error
	if extra.Repeat > 0 {
		_, err = s.Repository.MakeRepeatly(roomID, userName, startTimestamp, startTime, endTime, extra.Repeat, extra.Memo)
	} else {
		_, err = s.Repository.Make(roomID, userName, startTimestamp, startTime, endTime, extra.Memo)
	}
	return errors.WithStack(err)

}

func (s *Service) Cancel(reservationID int64) (bool, error) {
	return s.Repository.Cancel(reservationID)
}

type CustomTime time.Time

func (ct CustomTime) GetDateStr() string {
	y, m, d := time.Time(ct).Date()
	return fmt.Sprintf("%d-%02d-%02d", y, m, d)
}

func (ct CustomTime) GetHhmmInt() int {
	return time.Time(ct).Hour()*100 + time.Time(ct).Minute()
}
