package context

import (
	"context"

	"coverd/models"
)

type privateKey string

const (
	accountKey privateKey = "account"
)

func WithAccount(ctx context.Context, account *models.Account) context.Context {
	return context.WithValue(ctx, accountKey, account)
}

func Account(ctx context.Context) *models.Account {
	if temp := ctx.Value(accountKey); temp != nil {
		if account, ok := temp.(*models.Account); ok {
			return account
		}
	}
	return nil
}
