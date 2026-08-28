package files

import (
	"context"
	"errors"
	"net/http"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/session"
	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type IAMAuthorizationAdapter struct{ service *authorization.Service }

func NewIAMAuthorizationAdapter(db Database) (*IAMAuthorizationAdapter, error) {
	if db == nil {
		return nil, errors.New("files authorization database is required")
	}
	return &IAMAuthorizationAdapter{service: authorization.NewService(db)}, nil
}

func (adapter *IAMAuthorizationAdapter) RequireInTx(ctx context.Context, tx database.Tx, actorID, permission string) (Scope, error) {
	decision, err := adapter.service.RequireInTx(ctx, tx, actorID, permission)
	if errors.Is(err, authorization.ErrDenied) {
		return "", ErrDenied
	}
	if err != nil {
		return "", err
	}
	switch decision.Scope {
	case authorization.ScopeSelf:
		return ScopeSelf, nil
	case authorization.ScopeAll:
		return ScopeAll, nil
	default:
		return "", ErrDenied
	}
}

type SessionRequestService interface {
	AuthorizeRequest(context.Context, string, string, bool) (session.Issued, error)
}

type RequestIdentity struct {
	ActorID, CSRF     string
	ReplacementCookie *string
}

type IAMSessionRequestAdapter struct{ service SessionRequestService }

func NewIAMSessionRequestAdapter(service SessionRequestService) (*IAMSessionRequestAdapter, error) {
	if service == nil {
		return nil, errors.New("files session service is required")
	}
	return &IAMSessionRequestAdapter{service: service}, nil
}

func (*IAMSessionRequestAdapter) CookieName() string { return session.CookieName }

func (adapter *IAMSessionRequestAdapter) AuthorizeRequest(ctx context.Context, token, csrf string, mutation bool) (RequestIdentity, error) {
	issued, err := adapter.service.AuthorizeRequest(ctx, token, csrf, mutation)
	if err != nil {
		switch {
		case errors.Is(err, session.ErrAuthentication):
			return RequestIdentity{}, ErrAuthentication
		case errors.Is(err, session.ErrCSRF):
			return RequestIdentity{}, ErrCSRF
		default:
			return RequestIdentity{}, err
		}
	}
	identity := RequestIdentity{ActorID: issued.Profile.ID, CSRF: issued.CSRF}
	if issued.Rotated && issued.Token != "" {
		cookie := (&http.Cookie{Name: session.CookieName, Value: issued.Token, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}).String()
		identity.ReplacementCookie = &cookie
	}
	return identity, nil
}
