package organization

import (
	"context"

	iamadministration "go-admin/internal/modules/iam/administration"
	"go-admin/internal/platform/database"
)

type ProjectionAdapter struct {
	db         Database
	repository repository
}

func NewProjectionAdapter(db Database) (*ProjectionAdapter, error) {
	if db == nil {
		return nil, ErrInternal
	}
	return &ProjectionAdapter{db: db, repository: repository{dialect: db.Dialect()}}, nil
}

func (a *ProjectionAdapter) DepartmentLineage(ctx context.Context, id string) (iamadministration.OrganizationDepartmentLineage, error) {
	if a == nil || a.db == nil || id == "" {
		return iamadministration.OrganizationDepartmentLineage{}, ErrNotFound
	}
	result := iamadministration.OrganizationDepartmentLineage{DepartmentID: id, AncestorIDs: []string{}}
	err := a.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		departments, err := a.repository.departments(ctx, tx, false)
		if err != nil {
			return err
		}
		byID := make(map[string]Department, len(departments))
		for _, department := range departments {
			byID[department.ID] = department
		}
		current, ok := byID[id]
		if !ok {
			return ErrNotFound
		}
		seen := map[string]bool{id: true}
		for current.ParentID != nil {
			parentID := *current.ParentID
			if seen[parentID] {
				return ErrInternal
			}
			seen[parentID] = true
			result.AncestorIDs = append(result.AncestorIDs, parentID)
			current, ok = byID[parentID]
			if !ok {
				return ErrInternal
			}
		}
		return nil
	})
	if err != nil {
		return iamadministration.OrganizationDepartmentLineage{}, (&Service{db: a.db}).normalize(ctx, err)
	}
	return result, nil
}

func (a *ProjectionAdapter) PositionDepartment(ctx context.Context, id string) (string, error) {
	if a == nil || a.db == nil || id == "" {
		return "", ErrNotFound
	}
	var departmentID string
	err := a.db.WithinTx(ctx, func(ctx context.Context, tx database.Tx) error {
		position, err := a.repository.position(ctx, tx, id, false)
		if err != nil {
			return err
		}
		departmentID = position.DepartmentID
		return nil
	})
	if err != nil {
		return "", (&Service{db: a.db}).normalize(ctx, err)
	}
	return departmentID, nil
}

var _ iamadministration.OrganizationProjectionPort = (*ProjectionAdapter)(nil)
