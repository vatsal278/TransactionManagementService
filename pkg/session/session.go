package session

import "context"

type session struct{}

func SetSession(ctx context.Context, value any) context.Context {
	return context.WithValue(ctx, session{}, value)
}

func GetSession(ctx context.Context) any {
	return ctx.Value(session{})
}
