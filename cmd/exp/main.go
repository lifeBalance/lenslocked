package main

import (
	stdctx "context"
	"fmt"

	"github.com/lifebalance/lenslocked/context"
	"github.com/lifebalance/lenslocked/models"
)

func main() {
	ctx := stdctx.Background()
	user := models.User{
		Email: "bob@test.com",
	}
	ctx = context.WithUser(ctx, &user)

	// retrieve user from context
	retrievedUser := context.User(ctx)
	fmt.Println(retrievedUser.Email)
}
