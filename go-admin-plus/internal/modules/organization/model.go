// Package organization owns department trees, positions, and their persistence rules.
package organization

type Department struct {
	ID        string
	Key       string
	Name      string
	ParentID  *string
	SortOrder int
	Protected bool
}

type Position struct {
	ID           string
	Key          string
	Name         string
	DepartmentID string
	Enabled      bool
	Protected    bool
}

type PositionPage struct {
	Rows  []Position
	Total int
}

type DepartmentInput struct {
	Key, Name, ParentID string
	SortOrder           int
}

type PositionInput struct {
	Key, Name, DepartmentID string
	Enabled                 bool
}
