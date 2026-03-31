package db

import (
	"context"
	"errors"

	"github.com/goccy/go-json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*Queries
	ConnPool *pgxpool.Pool
}

func NewStore(connPool *pgxpool.Pool) *Store {
	return &Store{
		ConnPool: connPool,
		Queries:  New(connPool),
	}
}

type TxFn func(queries *Queries) error

// ExecTx executes a function within a database transaction
func (store *Store) ExecTx(ctx context.Context, fn TxFn) error {
	tx, err := store.ConnPool.Begin(ctx)
	if err != nil {
		return err
	}

	employeeID := employeeIDFromContext(ctx)

	if employeeID != uuid.Nil {
		_, err = tx.Exec(ctx, "SELECT set_config('myapp.current_employee_id', $1, true)", employeeID.String())
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				return rbErr
			}
			return err
		}
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}

type CreateEmployeeWithAccountTxParams struct {
	CreateUserParams     CreateUserParams
	CreateEmployeeParams CreateEmployeeProfileParams
	RoleID               uuid.UUID
}

type CreateEmployeeWithAccountTxResult struct {
	User     CustomUser
	Employee EmployeeProfile
}

func (store *Store) CreateEmployeeWithAccountTx(ctx context.Context, arg CreateEmployeeWithAccountTxParams) (CreateEmployeeWithAccountTxResult, error) {
	var result CreateEmployeeWithAccountTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		arg.CreateEmployeeParams.UserID = result.User.ID
		result.Employee, err = q.CreateEmployeeProfile(ctx, arg.CreateEmployeeParams)
		if err != nil {
			return err
		}

		// Auto-assign the active employee handbook for the employee's department (if configured).
		if arg.CreateEmployeeParams.DepartmentID != nil {
			template, tmplErr := q.GetActiveHandbookTemplateByDepartment(ctx, *arg.CreateEmployeeParams.DepartmentID)
			if tmplErr != nil {
				if !errors.Is(tmplErr, pgx.ErrNoRows) {
					return tmplErr
				}
			} else {
				assignedBy := ctx.Value("employee_id").(uuid.UUID)
				var assignedByPtr *uuid.UUID
				if assignedBy != uuid.Nil {
					assignedByPtr = &assignedBy
				}
				assigned, err := q.CreateEmployeeHandbookFromTemplate(ctx, CreateEmployeeHandbookFromTemplateParams{
					EmployeeID:           result.Employee.ID,
					TemplateID:           template.ID,
					AssignedByEmployeeID: assignedByPtr,
				})
				if err != nil {
					return err
				}

				_, err = q.CreateEmployeeHandbookAssignmentHistory(ctx, CreateEmployeeHandbookAssignmentHistoryParams{
					EmployeeHandbookID: &assigned.ID,
					EmployeeID:         result.Employee.ID,
					TemplateID:         template.ID,
					TemplateVersion:    assigned.TemplateVersion,
					Event:              HandbookAssignmentEventEnumAssigned,
					ActorEmployeeID:    assignedByPtr,
					Metadata: mustMarshalAssignmentMetadata(map[string]any{
						"source": "employee_creation",
					}),
				})
				if err != nil {
					return err
				}
			}
		}

		err = q.AssignRoleToUser(ctx, AssignRoleToUserParams{
			UserID: result.User.ID,
			RoleID: arg.RoleID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func mustMarshalAssignmentMetadata(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

func employeeIDFromContext(ctx context.Context) uuid.UUID {
	if ctx == nil {
		return uuid.Nil
	}

	value := ctx.Value("employee_id")
	switch v := value.(type) {
	case uuid.UUID:
		return v
	case *uuid.UUID:
		if v != nil {
			return *v
		}
	}

	return uuid.Nil
}
