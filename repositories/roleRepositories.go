package repositories

import (
	"context"
	"encoding/json"
	"it_school/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(conn *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{db: conn}
}

func (r *RoleRepository) GetRoleByID(c context.Context, roleID uuid.UUID) (*models.Role, error) {
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

func (r *RoleRepository) GetRoleByName(c context.Context, name string) (*models.Role, error) {
	var role models.Role
	var permissionsData []byte

	query := `SELECT id, name, permissions FROM roles WHERE name = $1`
	row := r.db.QueryRow(c, query, name)

	if err := row.Scan(&role.Id, &role.Name, &permissionsData); err != nil {
		return nil, err
	}

	if err := role.ScanPermissions(permissionsData); err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *models.Role) error {
    data, _ := json.Marshal(role.Permissions)
    _, err := r.db.Exec(ctx,
        `INSERT INTO roles (id, name, permissions) VALUES ($1, $2, $3)`,
        role.Id, role.Name, data)
    return err
}