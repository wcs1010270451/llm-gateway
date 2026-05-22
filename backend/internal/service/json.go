package service

import (
	"encoding/json"

	"gorm.io/datatypes"
)

func normalizeJSON(value map[string]any) (datatypes.JSON, error) {
	if value == nil {
		return datatypes.JSON([]byte("{}")), nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

func allowed(value string, values ...string) bool {
	for _, item := range values {
		if value == item {
			return true
		}
	}
	return false
}
