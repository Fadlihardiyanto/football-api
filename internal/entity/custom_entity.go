package entity

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type TimeOnly struct {
	time.Time
}

func (t *TimeOnly) Scan(value interface{}) error {
	if str, ok := value.(string); ok {
		parsed, err := time.Parse("15:04:05", str)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	}
	return fmt.Errorf("cannot scan %T into TimeOnly", value)
}

func (t TimeOnly) Value() (driver.Value, error) {
	return t.Format("15:04:05"), nil
}
