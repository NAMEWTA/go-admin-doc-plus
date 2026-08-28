package audit

import (
	"context"
	"errors"
	"net/http"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/session"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type SessionRequestService interface {
	AuthorizeRequest(context.Context, string, string, bool) (session.Issued, error)
}

// IAMRequestAuthorizer adapts the public Session request boundary without exposing credential
// material to Audit services, observers, or facts.
type IAMRequestAuthorizer struct{ sessions SessionRequestService }

func NewIAMRequestAuthorizer(sessions SessionRequestService) (*IAMRequestAuthorizer, error) {
	if sessions == nil {
		return nil, ErrInvalidArgument
	}
	return &IAMRequestAuthorizer{sessions: sessions}, nil
}

func (authorizer *IAMRequestAuthorizer) AuthorizeRequest(ctx context.Context, request *http.Request) (AuthorizedRequest, RequestFailure) {
	if authorizer == nil || authorizer.sessions == nil || request == nil {
		return AuthorizedRequest{}, RequestInternalFailed
	}
	token := ""
	if cookie, err := request.Cookie(session.CookieName); err == nil {
		token = cookie.Value
	}
	mutation := request.Method != http.MethodGet && request.Method != http.MethodHead
	issued, err := authorizer.sessions.AuthorizeRequest(ctx, token, request.Header.Get("X-CSRF-Token"), mutation)
	if errors.Is(err, session.ErrCSRF) {
		return AuthorizedRequest{}, RequestAuthorizationFailed
	}
	if errors.Is(err, session.ErrAuthentication) {
		return AuthorizedRequest{}, RequestAuthenticationFailed
	}
	if err != nil {
		return AuthorizedRequest{}, RequestInternalFailed
	}
	var replacement *string
	if issued.Rotated {
		value := (&http.Cookie{Name: session.CookieName, Value: issued.Token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}).String()
		replacement = &value
	}
	result, resultErr := NewAuthorizedRequest(Principal{ID: issued.Profile.ID}, issued.CSRF, replacement)
	if resultErr != nil {
		return AuthorizedRequest{}, RequestInternalFailed
	}
	return result, RequestAuthorized
}

// IAMPermissionAuthorizer keeps Audit's final decision in the same transaction as its query or
// cleanup operation, closing permission-revocation TOCTOU windows.
type IAMPermissionAuthorizer struct{ permissions *authorization.Service }

func NewIAMPermissionAuthorizer(permissions *authorization.Service) (*IAMPermissionAuthorizer, error) {
	if permissions == nil {
		return nil, ErrInvalidArgument
	}
	return &IAMPermissionAuthorizer{permissions: permissions}, nil
}

func (authorizer *IAMPermissionAuthorizer) Authorize(ctx context.Context, tx database.Tx, principal Principal, permission Permission) (AuthorizationDecision, error) {
	if authorizer == nil || authorizer.permissions == nil {
		return AuthorizationDenied, ErrInternal
	}
	_, err := authorizer.permissions.RequireInTx(ctx, tx, principal.ID, string(permission))
	if errors.Is(err, authorization.ErrDenied) {
		return AuthorizationDenied, nil
	}
	if err != nil {
		return AuthorizationDenied, err
	}
	return AuthorizationGranted, nil
}
