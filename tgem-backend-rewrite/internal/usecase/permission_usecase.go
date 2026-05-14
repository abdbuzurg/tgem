package usecase

import (
	"context"
	"errors"
	"fmt"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"

	"github.com/jackc/pgx/v5"
)

type permissionUsecase struct {
	q *db.Queries
}

func NewPermissionUsecase(q *db.Queries) IPermissionUsecase {
	return &permissionUsecase{q: q}
}

type IPermissionUsecase interface {
	GetAll() ([]model.Permission, error)
	GetByRoleName(roleName string) ([]dto.UserPermission, error)
	GetByRoleID(roleID uint) ([]model.Permission, error)
	GetEffectivePermissionsForUser(userID uint) ([]dto.EffectivePermission, error)
	GetByResourceURL(resourceURL string, roleID uint) error
	Create(data model.Permission) (model.Permission, error)
	CreateBatch(data []model.Permission) error
	Update(data model.Permission) (model.Permission, error)
	Delete(id uint) error
}

func (u *permissionUsecase) GetAll() ([]model.Permission, error) {
	rows, err := u.q.ListPermissions(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.Permission, len(rows))
	for i, r := range rows {
		out[i] = toModelPermission(r)
	}
	return out, nil
}

func (u *permissionUsecase) GetByRoleID(roleID uint) ([]model.Permission, error) {
	rows, err := u.q.ListPermissionsByRoleID(context.Background(), pgInt8(roleID))
	if err != nil {
		return nil, err
	}
	out := make([]model.Permission, len(rows))
	for i, r := range rows {
		out[i] = toModelPermission(r)
	}
	return out, nil
}

func (u *permissionUsecase) Create(data model.Permission) (model.Permission, error) {
	row, err := u.q.CreatePermission(context.Background(), db.CreatePermissionParams{
		RoleID:     pgInt8(data.RoleID),
		ResourceID: pgInt8(data.ResourceID),
		R:          pgBool(data.R),
		W:          pgBool(data.W),
		U:          pgBool(data.U),
		D:          pgBool(data.D),
	})
	if err != nil {
		return model.Permission{}, err
	}
	return toModelPermission(row), nil
}

func (u *permissionUsecase) Update(data model.Permission) (model.Permission, error) {
	row, err := u.q.UpdatePermission(context.Background(), db.UpdatePermissionParams{
		ID:         int64(data.ID),
		RoleID:     pgInt8(data.RoleID),
		ResourceID: pgInt8(data.ResourceID),
		R:          pgBool(data.R),
		W:          pgBool(data.W),
		U:          pgBool(data.U),
		D:          pgBool(data.D),
	})
	if err != nil {
		return model.Permission{}, err
	}
	return toModelPermission(row), nil
}

func (u *permissionUsecase) Delete(id uint) error {
	return u.q.DeletePermission(context.Background(), int64(id))
}

func (u *permissionUsecase) CreateBatch(data []model.Permission) error {
	rows := make([]db.CreatePermissionsBatchParams, len(data))
	for i, p := range data {
		rows[i] = db.CreatePermissionsBatchParams{
			RoleID:     pgInt8(p.RoleID),
			ResourceID: pgInt8(p.ResourceID),
			R:          pgBool(p.R),
			W:          pgBool(p.W),
			U:          pgBool(p.U),
			D:          pgBool(p.D),
		}
	}
	_, err := u.q.CreatePermissionsBatch(context.Background(), rows)
	return err
}

func (u *permissionUsecase) GetByRoleName(roleName string) ([]dto.UserPermission, error) {
	rows, err := u.q.ListUserPermissionsByRoleName(context.Background(), pgText(roleName))
	if err != nil {
		return nil, err
	}
	out := make([]dto.UserPermission, len(rows))
	for i, r := range rows {
		out[i] = dto.UserPermission{
			ResourceName: r.ResourceName.String,
			ResourceURL:  r.ResourceUrl.String,
			R:            r.R.Bool,
			W:            r.W.Bool,
			U:            r.U.Bool,
			D:            r.D.Bool,
		}
	}
	return out, nil
}

// GetEffectivePermissionsForUser returns the v2 flat permission list for the
// given user, suitable for the new /user/effective-permissions endpoint that
// the frontend declarative gate consumes.
func (u *permissionUsecase) GetEffectivePermissionsForUser(userID uint) ([]dto.EffectivePermission, error) {
	rows, err := u.q.ListEffectivePermissionsForUser(context.Background(), int64(userID))
	if err != nil {
		return nil, err
	}
	out := make([]dto.EffectivePermission, len(rows))
	for i, r := range rows {
		var pid *uint
		if r.ProjectID.Valid {
			v := uint(r.ProjectID.Int64)
			pid = &v
		}
		out[i] = dto.EffectivePermission{
			ProjectID:    pid,
			ResourceType: r.ResourceTypeCode,
			Action:       r.ActionCode,
		}
	}
	return out, nil
}

// GetByResourceURL preserves the GORM-era semantics exactly: missing rows
// are treated as "permitted" (returns nil); a row whose four flags are all
// false is "denied" (returns the Russian-language error). The pgx port
// folds the original errors.Is(err, gorm.ErrRecordNotFound) and the
// permission.ID == 0 fallthrough into a single pgx.ErrNoRows check, since
// sqlc's :one returns ErrNoRows for "no row" rather than (zero-value, nil).
func (u *permissionUsecase) GetByResourceURL(resourceURL string, roleID uint) error {
	permission, err := u.q.GetPermissionByResourceURL(context.Background(), db.GetPermissionByResourceURLParams{
		RoleID: pgInt8(roleID),
		Url:    pgText(resourceURL),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !permission.R.Bool && !permission.W.Bool && !permission.U.Bool && !permission.D.Bool {
		return fmt.Errorf("Доступ запрещен")
	}
	return nil
}

func toModelPermission(p db.Permission) model.Permission {
	return model.Permission{
		ID:         uint(p.ID),
		RoleID:     uintFromPgInt8(p.RoleID),
		ResourceID: uintFromPgInt8(p.ResourceID),
		R:          p.R.Bool,
		W:          p.W.Bool,
		U:          p.U.Bool,
		D:          p.D.Bool,
	}
}
