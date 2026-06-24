package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap 通用 JSON 字段类型
type JSONMap map[string]any

// Value 实现 driver.Valuer
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner
func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("JSONMap: unsupported type")
	}
	if len(data) == 0 {
		*j = nil
		return nil
	}
	return json.Unmarshal(data, j)
}

// JSONSlice JSON 数组
type JSONSlice []any

// Value 实现 driver.Valuer
func (j JSONSlice) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner
func (j *JSONSlice) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("JSONSlice: unsupported type")
	}
	if len(data) == 0 {
		*j = nil
		return nil
	}
	return json.Unmarshal(data, j)
}
