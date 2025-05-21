package utils

import "it_school/models"

func 	HasAccessToType(role *models.Role, typ string) bool {
	switch typ {
	case "урок":
		return role.Permissions["access_curator"] || role.Permissions["access_settings"]
	case "пролонгация", "заморозка":
		return role.Permissions["access_manager"] || role.Permissions["access_settings"]
	default:
		return false
	}
}