// utils/seed.go
package utils

import (
	"context"
	"fmt"
	"it_school/config"
	"it_school/logger"
	"it_school/models"
	"it_school/repositories"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdminAndRoles(rolesRepo *repositories.RoleRepository, usersRepo *repositories.UsersRepository) error {
  log := logger.GetLogger()
  c := context.Background()

  // --- 1) создаём (если ещё нет) три базовые роли ---
  needed := []struct {
      Name        string
      Permissions map[string]bool
  }{
      {Name: "admin",   Permissions: map[string]bool{"access_settings": true, "access_curator": true, "access_manager": true}},
      {Name: "manager", Permissions: map[string]bool{"access_settings": false,"access_curator": false,"access_manager": true}},
      {Name: "curator", Permissions: map[string]bool{"access_settings": false,"access_curator": true,"access_manager": false}},
  }

  for _, r := range needed {
      // пытаемся найти роль
      role, err := rolesRepo.GetRoleByName(c, r.Name)
      if err != nil {
          // не нашли — создаём
          role = &models.Role{
              Id:          uuid.New(),
              Name:        r.Name,
              Permissions: r.Permissions,
          }
          if err := rolesRepo.Create(c, role); err != nil {
              return fmt.Errorf("failed to create role %s: %w", r.Name, err)
          }
          log.Info("Created role", zap.String("role", r.Name))
      }
  }

  // --- 2) проверяем, есть ли хотя бы один админ, иначе создаём ---
  adminRole, err := rolesRepo.GetRoleByName(c, "admin")
  if err != nil {
      return fmt.Errorf("cannot lookup admin role after seeding: %w", err)
  }
  count, err := usersRepo.CountByRoleID(c, adminRole.Id)
  if err != nil {
      return fmt.Errorf("failed to count admin users: %w", err)
  }
  if count == 0 {
      // хешим пароль из env
      pwd := config.Config.Initial_Password
      hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
      user := &models.User{
          Id:           uuid.New(),
          Full_name:    config.Config.Admin_Name,
          Email:        config.Config.Admin_Mail,
          PasswordHash: string(hash),
          RoleID:       adminRole.Id,
          Telephone:    config.Config.Admin_Phone,
      }
      if _, err := usersRepo.Create(c, *user); err != nil {
          return fmt.Errorf("failed to create admin user: %w", err)
      }
      log.Info("Admin user created", zap.String("email", user.Email))
  }

  return nil
}

