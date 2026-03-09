package auth

import "context"

type ctxKey string

const (
	ctxUserID ctxKey = "auth_user_id"
	ctxEmail  ctxKey = "auth_email"
	ctxRole   ctxKey = "auth_role"
)

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxUserID, id)
}

func UserIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxUserID).(string)
	return id
}

func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, ctxEmail, email)
}

func EmailFromContext(ctx context.Context) string {
	email, _ := ctx.Value(ctxEmail).(string)
	return email
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, ctxRole, role)
}

func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ctxRole).(string)
	return role
}
