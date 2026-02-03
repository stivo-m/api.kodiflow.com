package helpers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ContextKey string

var (
	UserContextKey     ContextKey = "userId"
	BusinessContextKey ContextKey = "businessId"
)

func GetUserFromContext(ctx context.Context) (uuid.UUID, error) {
	storedValue := ctx.Value(UserContextKey)
	if storedValue == nil {
		return uuid.Nil, fmt.Errorf("no user found within the given context")
	}

	userIdString := storedValue.(string)

	userId, err := uuid.Parse(userIdString)
	if err != nil {
		return uuid.Nil, err
	}

	return userId, nil
}

func BusinessFromContext(ctx context.Context) (uuid.UUID, error) {
	storedValue := ctx.Value(BusinessContextKey)
	if storedValue == nil {
		return uuid.Nil, fmt.Errorf("no business found within the given context")
	}

	businessId := storedValue.(uuid.UUID)

	return businessId, nil
}
