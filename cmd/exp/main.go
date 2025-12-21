package main

import (
	"html/template"
	"os"
)

type User struct {
	Name string
	Age  int
	Meta UserMeta
}
type UserMeta struct {
	Visits int
}

func main() {
	t, err := template.ParseFiles("hello.gohtml")
	if err != nil {
		panic(err)
	}
	// Anonymous struct
	// user := struct {
	// 	Name string
	// 	Age  int
	// }{
	// 	Name: "Bob Sponge",
	// 	Age:  42,
	// }
	var user User
	user = User{
		Name: "Bob Sponge",
		Age:  42,
		Meta: UserMeta{
			Visits: 3,
		},
	}

	err = t.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
}
