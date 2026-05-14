package usecase

import (
	"context"
	"errors"
	"fmt"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/internal/security"
	"backend-v2/internal/utils"
	"backend-v2/model"
	"backend-v2/pkg/jwt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

// NewUserUsecase takes the pgx pool — Create and Update both wrap a
// transaction around the user-row write plus the user_in_projects
// management. The previous GORM signature took six repos
// (userRepo, userInProjectRepo, workerRepo, roleRepo, userInProjects, projectRepo);
// all that wiring is gone now.
func NewUserUsecase(pool *pgxpool.Pool) IUserUsecase {
	return &userUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IUserUsecase interface {
	GetAll() ([]model.User, error)
	GetPaginated(page, limit int, data model.User) ([]dto.UserPaginated, error)
	GetByID(id uint) (model.User, error)
	Create(data dto.NewUserData) error
	Update(data dto.NewUserData) error
	Delete(id uint) error
	Count() (int64, error)
	Login(data dto.LoginData) (dto.LoginResponse, error)
}

func (u *userUsecase) GetAll() ([]model.User, error) {
	rows, err := u.q.ListUsers(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.User, len(rows))
	for i, r := range rows {
		out[i] = toModelUser(r)
	}
	return out, nil
}

func (u *userUsecase) GetPaginated(page, limit int, data model.User) ([]dto.UserPaginated, error) {
	ctx := context.Background()

	var rows []db.User
	if !utils.IsEmptyFields(data) {
		filtered, err := u.q.ListUsersPaginatedFiltered(ctx, db.ListUsersPaginatedFilteredParams{
			Column1: int64(data.WorkerID),
			Column2: data.Username,
			Limit:   int32(limit),
			Offset:  int32((page - 1) * limit),
		})
		if err != nil {
			return []dto.UserPaginated{}, err
		}
		rows = filtered
	} else {
		paged, err := u.q.ListUsersPaginated(ctx, db.ListUsersPaginatedParams{
			Limit:  int32(limit),
			Offset: int32((page - 1) * limit),
		})
		if err != nil {
			return []dto.UserPaginated{}, err
		}
		rows = paged
	}

	var result []dto.UserPaginated
	for _, user := range rows {
		worker, err := u.q.GetWorker(ctx, int64(uintFromPgInt8(user.WorkerID)))
		if err != nil {
			return []dto.UserPaginated{}, err
		}

		if worker.Name.String == "Суперадмин" {
			continue
		}

		role, err := u.q.GetRole(ctx, int64(uintFromPgInt8(user.RoleID)))
		if err != nil {
			return []dto.UserPaginated{}, err
		}

		namesOfProjectsUserIn, err := u.q.ListProjectNamesByUserID(ctx, pgInt8(uint(user.ID)))
		if err != nil {
			return []dto.UserPaginated{}, err
		}

		result = append(result, dto.UserPaginated{
			ID:                 uint(user.ID),
			Username:           user.Username.String,
			WorkerName:         worker.Name.String,
			WorkerJobTitle:     worker.JobTitleInProject.String,
			WorkerMobileNumber: worker.MobileNumber.String,
			RoleName:           role.Name.String,
			AccessToProjects:   namesOfProjectsUserIn,
		})
	}

	return result, nil
}

func (u *userUsecase) GetByID(id uint) (model.User, error) {
	row, err := u.q.GetUser(context.Background(), int64(id))
	if err != nil {
		return model.User{}, err
	}
	return toModelUser(row), nil
}

func (u *userUsecase) Create(data dto.NewUserData) error {
	hashedPassword, err := security.Hash(data.UserData.Password)
	if err != nil {
		return err
	}
	data.UserData.Password = string(hashedPassword)

	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		WorkerID: pgInt8(data.UserData.WorkerID),
		Username: pgText(data.UserData.Username),
		Password: pgText(data.UserData.Password),
		RoleID:   pgInt8(data.UserData.RoleID),
	})
	if err != nil {
		return err
	}

	batch := make([]db.CreateUserInProjectsBatchParams, len(data.Projects))
	for i, projectID := range data.Projects {
		batch[i] = db.CreateUserInProjectsBatchParams{
			UserID:    pgInt8(uint(user.ID)),
			ProjectID: pgInt8(projectID),
		}
	}
	if len(batch) > 0 {
		if _, err := qtx.CreateUserInProjectsBatch(ctx, batch); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Update preserves the GORM-era password semantics exactly: an empty
// Password string keeps the existing hash (UPDATE uses
// COALESCE(NULLIF($4, ''), password)), a non-empty string overwrites it.
func (u *userUsecase) Update(data dto.NewUserData) error {
	var encryptedPassword string
	if data.UserData.Password != "" {
		hashedPassword, err := security.Hash(data.UserData.Password)
		if err != nil {
			return err
		}
		encryptedPassword = string(hashedPassword)
	}
	data.UserData.Password = encryptedPassword

	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.UpdateUser(ctx, db.UpdateUserParams{
		ID:       int64(data.UserData.ID),
		WorkerID: pgInt8(data.UserData.WorkerID),
		Username: pgText(data.UserData.Username),
		Column4:  data.UserData.Password,
		RoleID:   pgInt8(data.UserData.RoleID),
	}); err != nil {
		return err
	}

	if err := qtx.DeleteUserInProjectsByUserID(ctx, pgInt8(data.UserData.ID)); err != nil {
		return err
	}

	batch := make([]db.CreateUserInProjectsBatchParams, len(data.Projects))
	for i, projectID := range data.Projects {
		batch[i] = db.CreateUserInProjectsBatchParams{
			UserID:    pgInt8(data.UserData.ID),
			ProjectID: pgInt8(projectID),
		}
	}
	if len(batch) > 0 {
		if _, err := qtx.CreateUserInProjectsBatch(ctx, batch); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (u *userUsecase) Delete(id uint) error {
	return u.q.DeleteUser(context.Background(), int64(id))
}

func (u *userUsecase) Count() (int64, error) {
	return u.q.CountUsers(context.Background())
}

func (u *userUsecase) Login(data dto.LoginData) (dto.LoginResponse, error) {
	ctx := context.Background()

	user, err := u.q.GetUserByUsername(ctx, pgText(data.Username))
	if errors.Is(err, pgx.ErrNoRows) {
		return dto.LoginResponse{}, fmt.Errorf("Неправильное имя пользователя")
	}
	if err != nil {
		return dto.LoginResponse{}, err
	}

	if err := security.VerifyPassword(user.Password.String, data.Password); err != nil {
		return dto.LoginResponse{}, fmt.Errorf("Неправильный пароль")
	}

	// Superadmin bypass: a user whose role code is "superadmin" is treated
	// as having access to every project, so we skip the user_in_projects
	// check. The role code is set by migration 00005 (and seeded in
	// superadmin.sql) and is the same identity the v2 permission resolver
	// uses for the wildcard role grants.
	role, err := u.q.GetRole(ctx, int64(uintFromPgInt8(user.RoleID)))
	if err != nil {
		return dto.LoginResponse{}, err
	}

	if role.Code != "superadmin" {
		userInProjects, err := u.q.ListUserInProjectsByUserID(ctx, pgInt8(uint(user.ID)))
		if err != nil {
			return dto.LoginResponse{}, fmt.Errorf("У вас нету доступа в выбранный проект")
		}

		access := false
		for _, uip := range userInProjects {
			if uintFromPgInt8(uip.ProjectID) == data.ProjectID {
				access = true
				break
			}
		}

		if !access {
			return dto.LoginResponse{}, fmt.Errorf("У вас нету доступа в выбранный проект")
		}
	}

	result := dto.LoginResponse{Admin: false}
	token, err := jwt.CreateToken(uint(user.ID), uintFromPgInt8(user.WorkerID), uintFromPgInt8(user.RoleID), data.ProjectID)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	result.Token = token

	project, err := u.q.GetProject(ctx, int64(data.ProjectID))
	if err != nil {
		return dto.LoginResponse{}, err
	}

	if project.Name.String == "Администрирование" {
		result.Admin = true
	}

	return result, nil
}

func toModelUser(u db.User) model.User {
	return model.User{
		ID:       uint(u.ID),
		WorkerID: uintFromPgInt8(u.WorkerID),
		Username: u.Username.String,
		Password: u.Password.String,
		RoleID:   uintFromPgInt8(u.RoleID),
	}
}
