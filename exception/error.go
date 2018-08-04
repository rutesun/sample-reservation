package exception

import "github.com/pkg/errors"

var Unavailable = errors.New("예약이 불가능합니다.")
