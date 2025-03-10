package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const TimeFormat = "2006-01-02"

type LocalDate time.Time

func (t *LocalDate) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 2 {
		*t = LocalDate(time.Time{})
		return
	}

	now, err := time.Parse(`"`+TimeFormat+`"`, string(data))
	*t = LocalDate(now)
	return
}
func (t LocalDate) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = tt.AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

func (t LocalDate) Value() (driver.Value, error) {
	if t.String() == "0001-01-01" {
		return nil, nil
	}
	if t.IsZero() {
		return nil, nil // 返回 nil 表示数据库中的 NULL
	}
	return []byte(time.Time(t).Format(TimeFormat)), nil
}

func (t *LocalDate) Scan(v interface{}) error {
	if v == nil {
		*t = LocalDate(time.Time{})
		return nil
	}

	// 尝试解析时间
	switch v := v.(type) {
	case time.Time:
		*t = LocalDate(v)
	case string:
		tTime, err := time.Parse(TimeFormat, v)
		if err != nil {
			return err
		}
		*t = LocalDate(tTime)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}

func (t LocalDate) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t LocalDate) String() string {
	if t.IsZero() {
		return ""
	}
	return time.Time(t).Format(TimeFormat)
}
