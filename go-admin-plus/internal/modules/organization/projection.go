package organization

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type DepartmentLineage struct {
	DepartmentID string
	AncestorIDs  []string
}

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

func (a *ProjectionAdapter) DepartmentLineage(ctx context.Context, id string) (DepartmentLineage, error) {
	if a == nil || a.db == nil || id == "" {
		return DepartmentLineage{}, ErrNotFound
	}
	result := DepartmentLineage{DepartmentID: id, AncestorIDs: []string{}}
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
		return DepartmentLineage{}, (&Service{db: a.db}).normalize(ctx, err)
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

func (a *ProjectionAdapter) PositionDepartmentInTx(ctx context.Context, tx database.Tx, id string) (string, error) {
	if a == nil || tx == nil || id == "" {
		return "", ErrNotFound
	}
	position, err := a.repository.position(ctx, tx, id, true)
	if err != nil || !position.Enabled {
		return "", ErrNotFound
	}
	return position.DepartmentID, nil
}

func (a *ProjectionAdapter) DepartmentSetInTx(ctx context.Context, tx database.Tx, id string, descendants bool) ([]string, error) {
	if a == nil || tx == nil || id == "" {
		return nil, ErrNotFound
	}
	departments, err := a.repository.departments(ctx, tx, true)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]Department, len(departments))
	children := make(map[string][]string)
	for _, department := range departments {
		byID[department.ID] = department
		if department.ParentID != nil {
			children[*department.ParentID] = append(children[*department.ParentID], department.ID)
		}
	}
	if _, ok := byID[id]; !ok {
		return nil, ErrNotFound
	}
	result, seen, queue := []string{}, map[string]bool{}, []string{id}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if seen[current] {
			return nil, ErrInternal
		}
		seen[current] = true
		result = append(result, current)
		if descendants {
			queue = append(queue, children[current]...)
		}
	}
	return result, nil
}

func (a *ProjectionAdapter) ValidateDepartmentIDsInTx(ctx context.Context, tx database.Tx, ids []string) error {
	if a == nil || tx == nil {
		return ErrNotFound
	}
	departments, err := a.repository.departments(ctx, tx, true)
	if err != nil {
		return err
	}
	known := make(map[string]struct{}, len(departments))
	for _, department := range departments {
		known[department.ID] = struct{}{}
	}
	for _, id := range ids {
		if _, ok := known[id]; !ok {
			return ErrNotFound
		}
	}
	return nil
}
