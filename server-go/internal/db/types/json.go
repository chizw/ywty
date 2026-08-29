// Package types 提供 GORM 的自定义列类型。
package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON 表示 JSON 列：MySQL 为 JSON 类型、SQLite 为 TEXT，
// 底层以 JSON 文本存取，空值写 NULL。
type JSON json.RawMessage

// String 构造：把任意值 JSON 序列化为 JSON 列值。
func NewJSON(v any) (JSON, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return JSON(b), nil
}

// MustJSON 同 NewJSON，序列化失败时 panic（仅用于构造期常量）。
func MustJSON(v any) JSON {
	j, err := NewJSON(v)
	if err != nil {
		panic(err)
	}
	return j
}

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = JSON(v)
	default:
		return fmt.Errorf("types: 不支持的 JSON 列类型 %T", value)
	}
	return nil
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

func (j *JSON) UnmarshalJSON(b []byte) error {
	*j = append((*j)[0:0], b...)
	return nil
}

// Any 把列值反序列化到任意 Go 类型。
func (j JSON) Any(dst any) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, dst)
}

// IsEmpty 表示列为 NULL 或空。
func (j JSON) IsEmpty() bool {
	return len(j) == 0 || string(j) == "null"
}

// GormDataType 声明 GORM 通用类型。
func (JSON) GormDataType() string { return "json" }

var _ driver.Valuer = (*JSON)(nil)
var _ sql.Scanner = (*JSON)(nil)
