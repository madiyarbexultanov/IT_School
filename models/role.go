package models

import "encoding/json"

type Role struct {
	Id          int             `json:"id"`
	Name        string          `json:"name"`
	Permissions map[string]bool `json:"permissions"`
}


func (r *Role) ScanPermissions(data []byte) error {
	return json.Unmarshal(data, &r.Permissions)
}
