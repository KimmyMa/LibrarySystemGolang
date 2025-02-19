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
	fmt.Println(tt)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	b := make([]byte, 0, len(TimeFormat)+2)
	b = append(b, '"')
	b = tt.AppendFormat(b, TimeFormat)
	b = append(b, '"')
	return b, nil
}

//func (t LocalDate) MarshalJSON() ([]byte, error) {
//	b := make([]byte, 0, len(TimeFormat)+2)
//	b = append(b, '"')
//	b = time.Time(t).AppendFormat(b, TimeFormat)
//	b = append(b, '"')
//	return b, nil
//}

func (t LocalDate) Value() (driver.Value, error) {
	if t.String() == "0001-01-01" {
		return nil, nil
	}
	return []byte(time.Time(t).Format(TimeFormat)), nil
}

//	func (t *LocalDate) Scan(v interface{}) error {
//		if v == nil {
//			*t = LocalDate(time.Time{})
//			fmt.Println("v is nil, t is ", t)
//			return nil
//		}
//		tTime, err := time.Parse("2006-01-02 15:04:05 +0800 CST", v.(time.Time).String())
//		if err != nil {
//			return err
//		}
//		*t = LocalDate(tTime)
//		fmt.Println(t)
//		return nil
//	}
func (t *LocalDate) Scan(v interface{}) error {
	if v == nil {
		*t = LocalDate(time.Time{})
		//fmt.Println("Scanned NULL value")
		return nil
	}

	// 尝试解析时间
	switch v := v.(type) {
	case time.Time:
		*t = LocalDate(v)
		//fmt.Println("Scanned time.Time value:", v)
	case string:
		tTime, err := time.Parse(TimeFormat, v)
		if err != nil {
			return err
		}
		*t = LocalDate(tTime)
		//fmt.Println("Scanned string value:", v)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}
func (t LocalDate) String() string {
	if time.Time(t).Year() == 1 {
		return ""
	}
	return time.Time(t).Format(TimeFormat)
}
