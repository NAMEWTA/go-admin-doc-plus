package administration

import "context"

// OrganizationProjectionPort is the IAM-owned view required to resolve future organization scope.
// Implementations must return a stable not-found error for unknown identifiers and no persistence records.
type OrganizationProjectionPort interface {
	DepartmentLineage(context.Context, string) (OrganizationDepartmentLineage, error)
	PositionDepartment(context.Context, string) (string, error)
}

type OrganizationDepartmentLineage struct {
	DepartmentID string
	AncestorIDs  []string
}
