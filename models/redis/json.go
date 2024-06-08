package redis

import (
	"strings"
	"time"
)

type JsonTime struct {
	time.Time
}

func (p *JsonTime) UnmarshalJSON(data []byte) error {
	dataStr := string(data)
	if data != nil && dataStr != "null" && len(data) > 0 {
		var local time.Time
		var err error
		if strings.HasSuffix(dataStr, "+0000\"") {
			local, err = time.ParseInLocation("\"2006-01-02T15:04:05.000+0000\"", dataStr, time.Local)
			if err != nil {
				err = (&local).UnmarshalJSON(data)
			}
		} else {
			err = (&local).UnmarshalJSON(data)
		}
		*p = JsonTime{Time: local}
		return err
	} else {
		return nil
	}
}
func (c *JsonTime) MarshalJSON() ([]byte, error) {
	if &c.Time != nil {
		data := make([]byte, 0)
		data = append(data, '"')
		data = time.Time(c.Time).AppendFormat(data, "2006-01-02 15:04:05.000+0000")
		data = append(data, '"')
		return data, nil
	} else {
		return nil, nil
	}
}
