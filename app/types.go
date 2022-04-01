package app

import (
	"encoding/json"
	"fmt"
	"time"
)

const DateTimeFormat = "2006-01-02T15:04:05.000000Z07:00"

type DateTime struct {
	time.Time
}

func (dt DateTime) String() string {
	return dt.Time.Format(DateTimeFormat)
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.Format(DateTimeFormat))
}

func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	t, err := time.Parse(DateTimeFormat, str)
	if err != nil {
		return err
	}
	*dt = DateTime{t}
	return nil
}

func (dt *DateTime) RedisScan(src interface{}) (err error) {
	if dt == nil {
		return fmt.Errorf("nil pointer")
	}
	switch src := src.(type) {
	case string:
		dt.Time, err = time.Parse(DateTimeFormat, src)
	case []uint8:
		dt.Time, err = time.Parse(DateTimeFormat, string(src))
	default:
		err = fmt.Errorf("cannot convert from %T to %T", src, dt)
	}
	return err
}
