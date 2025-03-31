package repositories

import (
	"context"
	"it_school/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(conn *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: conn}
}

func (r *RoleRepository) GetRoleByID(c context.Context, roleID int) (*models.Role, error) {
	var role models.Role
	var permissionsData []byte
	query := `SELECT id, name, permissions FROM roles WHERE id = $1`
	row := r.db.QueryRow(c, query, roleID)

	if err := row.Scan(&role.Id, &role.Name, &permissionsData); err != nil {
		return nil, err
	}

	if err := role.ScanPermissions(permissionsData); err != nil {
		return nil, err
	}

	return &role, nil
}
