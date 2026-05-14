package usecase

import (
	"context"

	"backend-v2/internal/db"
	"backend-v2/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type projectUsecase struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewProjectUsecase(pool *pgxpool.Pool) IProjectUsecase {
	return &projectUsecase{
		pool: pool,
		q:    db.New(pool),
	}
}

type IProjectUsecase interface {
	GetAll() ([]model.Project, error)
	GetPaginated(page, limit int) ([]model.Project, error)
	GetByID(id uint) (model.Project, error)
	Create(data model.Project) (model.Project, error)
	Update(data model.Project) (model.Project, error)
	Delete(id uint) error
	Count() (int64, error)
	GetProjectName(projectID uint) (string, error)
}

func (u *projectUsecase) GetAll() ([]model.Project, error) {
	rows, err := u.q.ListProjects(context.Background())
	if err != nil {
		return nil, err
	}
	out := make([]model.Project, len(rows))
	for i, r := range rows {
		out[i] = toModelProject(r)
	}
	return out, nil
}

func (u *projectUsecase) GetPaginated(page, limit int) ([]model.Project, error) {
	rows, err := u.q.ListProjectsExcludeAdmin(context.Background(), db.ListProjectsExcludeAdminParams{
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.Project, len(rows))
	for i, r := range rows {
		out[i] = toModelProject(r)
	}
	return out, nil
}

func (u *projectUsecase) GetByID(id uint) (model.Project, error) {
	row, err := u.q.GetProject(context.Background(), int64(id))
	if err != nil {
		return model.Project{}, err
	}
	return toModelProject(row), nil
}

// Create wraps a transaction: insert the project, link user 1 (the
// initial superadmin) into UserInProject, and seed four InvoiceCount
// rows (one per invoice flavor that has a counter). Errors from any
// step roll back the entire transaction.
func (u *projectUsecase) Create(data model.Project) (model.Project, error) {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return model.Project{}, err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	row, err := qtx.CreateProject(ctx, db.CreateProjectParams{
		Name:                 pgText(data.Name),
		Client:               pgText(data.Client),
		Budget:               pgNumericFromDecimal(data.Budget),
		BudgetCurrency:       pgText(data.BudgetCurrency),
		Description:          pgText(data.Description),
		SignedDateOfContract: pgTimestamptz(data.SignedDateOfContract),
		DateStart:            pgTimestamptz(data.DateStart),
		DateEnd:              pgTimestamptz(data.DateEnd),
		ProjectManager:       pgText(data.ProjectManager),
	})
	if err != nil {
		return model.Project{}, err
	}

	if err := qtx.CreateUserInProject(ctx, db.CreateUserInProjectParams{
		UserID:    pgInt8(1),
		ProjectID: pgInt8(uint(row.ID)),
	}); err != nil {
		return model.Project{}, err
	}

	for _, t := range []string{"input", "output", "return", "writeoff", "output-out-of-project"} {
		if err := qtx.CreateInvoiceCount(ctx, db.CreateInvoiceCountParams{
			ProjectID:   pgInt8(uint(row.ID)),
			InvoiceType: pgText(t),
			Count:       pgInt8(0),
		}); err != nil {
			return model.Project{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Project{}, err
	}
	return toModelProject(row), nil
}

func (u *projectUsecase) Update(data model.Project) (model.Project, error) {
	row, err := u.q.UpdateProject(context.Background(), db.UpdateProjectParams{
		ID:                   int64(data.ID),
		Name:                 pgText(data.Name),
		Client:               pgText(data.Client),
		Budget:               pgNumericFromDecimal(data.Budget),
		BudgetCurrency:       pgText(data.BudgetCurrency),
		Description:          pgText(data.Description),
		SignedDateOfContract: pgTimestamptz(data.SignedDateOfContract),
		DateStart:            pgTimestamptz(data.DateStart),
		DateEnd:              pgTimestamptz(data.DateEnd),
		ProjectManager:       pgText(data.ProjectManager),
	})
	if err != nil {
		return model.Project{}, err
	}
	return toModelProject(row), nil
}

// Delete wraps a transaction: remove the user_in_projects rows for the
// project, then the project row itself. Errors roll back the tx.
func (u *projectUsecase) Delete(id uint) error {
	ctx := context.Background()
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)
	if err := qtx.DeleteUserInProjectsByProjectID(ctx, pgInt8(id)); err != nil {
		return err
	}
	if err := qtx.DeleteProject(ctx, int64(id)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (u *projectUsecase) Count() (int64, error) {
	return u.q.CountProjects(context.Background())
}

func (u *projectUsecase) GetProjectName(projectID uint) (string, error) {
	return u.q.GetProjectName(context.Background(), int64(projectID))
}

func toModelProject(p db.Project) model.Project {
	return model.Project{
		ID:                   uint(p.ID),
		Name:                 p.Name.String,
		Client:               p.Client.String,
		Budget:               decimalFromPgNumeric(p.Budget),
		BudgetCurrency:       p.BudgetCurrency.String,
		Description:          p.Description.String,
		SignedDateOfContract: timeFromPgTimestamptz(p.SignedDateOfContract),
		DateStart:            timeFromPgTimestamptz(p.DateStart),
		DateEnd:              timeFromPgTimestamptz(p.DateEnd),
		ProjectManager:       p.ProjectManager.String,
	}
}
