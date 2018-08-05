package reservation

import (
	"time"

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
	CheckAvailable(roomID int, startTime time.Time, endTime time.Time) (bool, error)
	Make(roomID int, userID int, startTime time.Time, endTime time.Time, memo string) error
	MakeRepeatly(roomID int, userID int, startTime time.Time, endTime time.Time, repeatCnt int, memo string) error
	Cancel(reservationID uint) (bool, error)
}

type service struct {
	Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) List(date time.Time) (ReservedMap, error) {
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
func (s *service) CheckAvailable(roomID int, startTime time.Time, endTime time.Time) (bool, error) {
	if startTime.Before(endTime) {
		return false, exception.InvalidRequest
	}

	return s.Repository.CheckAvailable(roomID, startTime, endTime)
}

func (s *service) Make(roomID int, userID int, startTime time.Time, endTime time.Time, extra ExtraInfo) error {
	if extra.Repeat > 0 {
		return s.Repository.MakeRepeatly(roomID, userID, startTime, endTime, extra.Repeat, extra.Memo)
	}
	return s.Repository.Make(roomID, userID, startTime, endTime, extra.Memo)
}

func (s *service) Cancel(reservationID uint) (bool, error) {
	return s.Repository.Cancel(reservationID)
}
