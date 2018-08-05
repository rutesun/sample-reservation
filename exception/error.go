package exception

import "github.com/pkg/errors"

var (
	Unavailable      = errors.New("예약이 불가능합니다")
	InvalidCondition = errors.New("잘못된 요청입니다")
	InvalidRequest   = errors.New("잘못된 요청입니다")
)
