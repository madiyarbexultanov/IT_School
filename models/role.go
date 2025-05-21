package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Role struct {
	Id          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Permissions map[string]bool `json:"permissions"`
}


func (r *Role) ScanPermissions(data []byte) error {
	return json.Unmarshal(data, &r.Permissions)
}
