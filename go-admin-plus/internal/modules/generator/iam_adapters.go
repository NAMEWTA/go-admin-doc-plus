package generator

import (
	"context"
	"errors"
	"net/http"

	"go-admin/internal/modules/iam/authorization"
	"go-admin/internal/modules/iam/session"
)

type IAMAuthorizationAdapter struct{ service *authorization.Service }

func NewIAMAuthorizationAdapter(db authorization.Database) (*IAMAuthorizationAdapter, error) {
	if db == nil {
		return nil, ErrInvalid
	}
	return &IAMAuthorizationAdapter{service: authorization.NewService(db)}, nil
}

func (adapter *IAMAuthorizationAdapter) Require(ctx context.Context, actorID, permission string) error {
	_, err := adapter.service.Require(ctx, actorID, permission)
	if errors.Is(err, authorization.ErrDenied) {
		return ErrDenied
	}
	return err
}

type SessionRequestService interface {
	AuthorizeRequest(context.Context, string, string, bool) (session.Issued, error)
}
type IAMSessionRequestAdapter struct{ service SessionRequestService }

func NewIAMSessionRequestAdapter(service SessionRequestService) (*IAMSessionRequestAdapter, error) {
	if service == nil {
		return nil, ErrInvalid
	}
	return &IAMSessionRequestAdapter{service: service}, nil
}
func (*IAMSessionRequestAdapter) CookieName() string { return session.CookieName }
func (adapter *IAMSessionRequestAdapter) AuthorizeRequest(ctx context.Context, token, csrf string, mutation bool) (RequestIdentity, error) {
	grant, err := adapter.service.AuthorizeRequest(ctx, token, csrf, mutation)
	if err != nil {
		switch {
		case errors.Is(err, session.ErrCSRF):
			return RequestIdentity{}, ErrCSRF
		case errors.Is(err, session.ErrAuthentication):
			return RequestIdentity{}, ErrAuthentication
		default:
			return RequestIdentity{}, err
		}
	}
	identity := RequestIdentity{ActorID: grant.Profile.ID, CSRF: grant.CSRF}
	if grant.Rotated && grant.Token != "" {
		cookie := (&http.Cookie{Name: session.CookieName, Value: grant.Token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}).String()
		identity.ReplacementCookie = &cookie
	}
	return identity, nil
}
