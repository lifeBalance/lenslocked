package main

import (
	"fmt"
)

type ctxKey string // Don't export it (package scope)

const (
	// Don't export it (package scope)ß
	favouriteColorKey ctxKey = "fav-col"
)

func main() {
	// ctx := context.Background()
	// fmt.Println(ctx)
	// ctx = context.WithValue(ctx, favouriteColorKey, "blue")
	// ctx = context.WithValue(ctx, "fav-num", 42)
	// fmt.Println(ctx)

	// Get and print a value
	// value := ctx.Value("fav-col")
	// fmt.Println(value) // nil

	// Get and print another value
	// value = ctx.Value(favouriteColorKey)
	// fmt.Println(value) // blue

	// Get and print another value
	// value = ctx.Value("fav-num")
	// fmt.Println(value) // 42

	var i interface{} = "hello"
	fmt.Println(i) // hello
	// Gets the value if it's a string
	strValue, ok := i.(string)
	if ok {
		fmt.Println(strValue) // hello
	}
	// intValue := i.(int) // panic (assertion fails bc no ok used)

	// Gets the value if it's an int
	intValue, ok := i.(int)
	if ok {
		fmt.Println(intValue)
	} else {
		fmt.Println("not an int") // not an int
	}

	i = 42
	intValue, ok = i.(int)
	if ok {
		fmt.Println(intValue) // 42
	} else {
		fmt.Println("not an int")
	}
	fmt.Println(i) // 42
}
