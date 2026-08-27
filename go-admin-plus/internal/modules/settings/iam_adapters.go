package settings

import (
	"context"
	"errors"
	"net/http"

	"go-admin/internal/modules/iam/authorization"
	"go-admin/internal/modules/iam/session"
	"go-admin/internal/platform/database"
)

type IAMAuthorizationAdapter struct {
	service *authorization.Service
	dialect database.Dialect
}

func NewIAMAuthorizationAdapter(db Database) (*IAMAuthorizationAdapter, error) {
	if db == nil {
		return nil, errors.New("settings authorization database is required")
	}
	return &IAMAuthorizationAdapter{service: authorization.NewService(db), dialect: db.Dialect()}, nil
}
func (a *IAMAuthorizationAdapter) Dialect() database.Dialect { return a.dialect }
func (a *IAMAuthorizationAdapter) RequireInTx(ctx context.Context, tx database.Tx, actorID, permission string) (Scope, error) {
	decision, err := a.service.RequireInTx(ctx, tx, actorID, permission)
	if errors.Is(err, authorization.ErrDenied) {
		return "", ErrDenied
	}
	if err != nil {
		return "", err
	}
	if decision.Scope != authorization.ScopeAll {
		return "", ErrDenied
	}
	return ScopeAll, nil
}

type SessionRequestService interface {
	AuthorizeRequest(context.Context, string, string, bool) (session.Issued, error)
}
type IAMSessionRequestAdapter struct{ service SessionRequestService }

func NewIAMSessionRequestAdapter(service SessionRequestService) (*IAMSessionRequestAdapter, error) {
	if service == nil {
		return nil, errors.New("settings session service is required")
	}
	return &IAMSessionRequestAdapter{service: service}, nil
}
func (*IAMSessionRequestAdapter) CookieName() string { return session.CookieName }
func (a *IAMSessionRequestAdapter) AuthorizeRequest(ctx context.Context, token, csrf string, mutation bool) (RequestIdentity, error) {
	grant, err := a.service.AuthorizeRequest(ctx, token, csrf, mutation)
	if err != nil {
		if errors.Is(err, session.ErrCSRF) {
			return RequestIdentity{}, ErrCSRF
		}
		if errors.Is(err, session.ErrAuthentication) {
			return RequestIdentity{}, ErrAuthentication
		}
		return RequestIdentity{}, err
	}
	result := RequestIdentity{ActorID: grant.Profile.ID, CSRF: grant.CSRF}
	if grant.Rotated && grant.Token != "" {
		value := (&http.Cookie{Name: session.CookieName, Value: grant.Token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}).String()
		result.ReplacementCookie = &value
	}
	return result, nil
}
