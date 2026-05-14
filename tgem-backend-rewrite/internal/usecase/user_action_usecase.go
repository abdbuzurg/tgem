package usecase

import (
	"context"
	"fmt"
	"time"

	"backend-v2/internal/db"
	"backend-v2/internal/dto"
	"backend-v2/model"
)

type userActionUsecase struct {
	q *db.Queries
}

func NewUserActionUsecase(q *db.Queries) IUserActionUsecase {
	return &userActionUsecase{q: q}
}

type IUserActionUsecase interface {
	GetAllByUserID(userID uint) ([]dto.UserActionView, error)
	GetPaginated(filter dto.UserActionFilter) ([]dto.UserActionView, int64, error)
	ListFilterUserOptions() ([]dto.UserActionFilterUserOption, error)
	Create(data model.UserAction)
}

func (u *userActionUsecase) GetAllByUserID(userID uint) ([]dto.UserActionView, error) {
	rows, err := u.q.ListUserActionsByUserID(context.Background(), pgInt8(userID))
	if err != nil {
		return []dto.UserActionView{}, err
	}

	result := make([]dto.UserActionView, len(rows))
	for i, r := range rows {
		result[i] = dto.UserActionView{
			ID:                  uint(r.ID),
			ActionType:          r.ActionType.String,
			ActionID:            uintFromPgInt8(r.ActionID),
			ActionStatus:        r.ActionStatus.Bool,
			ActionStatusMessage: r.ActionStatusMessage.String,
			ActionURL:           r.ActionUrl.String,
			HTTPMethod:          r.HttpMethod.String,
			RequestIP:           r.RequestIp.String,
			UserID:              uintFromPgInt8(r.UserID),
			ProjectID:           uintFromPgInt8(r.ProjectID),
			DateOfAction:        timeFromPgTimestamptz(r.DateOfAction),
		}
	}

	return result, nil
}

// GetPaginated returns the page-N slice of user actions matching the filter
// plus the total count for the same WHERE clause (so the UI can paginate).
// Filter fields left at their zero value are interpreted as "no filter".
func (u *userActionUsecase) GetPaginated(filter dto.UserActionFilter) ([]dto.UserActionView, int64, error) {
	ctx := context.Background()

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 25
	}

	// Open-ended date sentinels: filter.DateFrom zero → no lower bound,
	// filter.DateTo zero → no upper bound. Postgres won't reject 0001-01-01
	// or 9999-12-31 from a timestamptz literal.
	dateFrom := filter.DateFrom
	if dateFrom.IsZero() {
		dateFrom = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	dateTo := filter.DateTo
	if dateTo.IsZero() {
		dateTo = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	statusValue := pgBool(false)
	if filter.HasStatus {
		statusValue = pgBool(filter.Status)
	}

	listParams := db.ListUserActionsPaginatedParams{
		Column1:        int64(filter.UserID),
		Column2:        int64(filter.ProjectID),
		Column3:        filter.ActionType,
		Column4:        filter.HTTPMethod,
		Column5:        filter.HasStatus,
		ActionStatus:   statusValue,
		DateOfAction:   pgTimestamptz(dateFrom),
		DateOfAction_2: pgTimestamptz(dateTo),
		Limit:          int32(filter.Limit),
		Offset:         int32((filter.Page - 1) * filter.Limit),
	}

	rows, err := u.q.ListUserActionsPaginated(ctx, listParams)
	if err != nil {
		return nil, 0, err
	}

	count, err := u.q.CountUserActionsPaginated(ctx, db.CountUserActionsPaginatedParams{
		Column1:        listParams.Column1,
		Column2:        listParams.Column2,
		Column3:        listParams.Column3,
		Column4:        listParams.Column4,
		Column5:        listParams.Column5,
		ActionStatus:   listParams.ActionStatus,
		DateOfAction:   listParams.DateOfAction,
		DateOfAction_2: listParams.DateOfAction_2,
	})
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserActionView, len(rows))
	for i, r := range rows {
		result[i] = dto.UserActionView{
			ID:                  uint(r.ID),
			ActionURL:           r.ActionUrl,
			ActionType:          r.ActionType,
			ActionID:            uint(r.ActionID),
			ActionStatus:        r.ActionStatus,
			ActionStatusMessage: r.ActionStatusMessage,
			HTTPMethod:          r.HttpMethod,
			RequestIP:           r.RequestIp,
			UserID:              uint(r.UserID),
			ProjectID:           uint(r.ProjectID),
			Username:            r.Username,
			DateOfAction:        timeFromPgTimestamptz(r.DateOfAction),
		}
	}

	return result, count, nil
}

// ListFilterUserOptions returns a flat (id, username, workerName) list used
// by the audit-log page's filter dropdown. The list is small (one row per
// app user) and stable, so we don't paginate it.
func (u *userActionUsecase) ListFilterUserOptions() ([]dto.UserActionFilterUserOption, error) {
	rows, err := u.q.ListUserOptionsForAudit(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserActionFilterUserOption, len(rows))
	for i, r := range rows {
		result[i] = dto.UserActionFilterUserOption{
			ID:         uint(r.ID),
			Username:   r.Username,
			WorkerName: r.WorkerName,
		}
	}
	return result, nil
}

// Create writes a user_action best-effort; errors are logged and
// swallowed to mirror the GORM-era behavior (the audit log shouldn't
// abort the action it's recording).
func (u *userActionUsecase) Create(data model.UserAction) {
	row, err := u.q.CreateUserAction(context.Background(), db.CreateUserActionParams{
		ActionUrl:           pgText(data.ActionURL),
		ActionType:          pgText(data.ActionType),
		ActionID:            pgInt8(data.ActionID),
		ActionStatus:        pgBool(data.ActionStatus),
		ActionStatusMessage: pgText(data.ActionStatusMessage),
		HttpMethod:          pgText(data.HTTPMethod),
		RequestIp:           pgText(data.RequestIP),
		UserID:              pgInt8(data.UserID),
		ProjectID:           pgInt8(data.ProjectID),
		DateOfAction:        pgTimestamptz(data.DateOfAction),
	})
	if err != nil {
		fmt.Printf("could not save user action - %+v \n with error - %v", row, err)
	}
}
