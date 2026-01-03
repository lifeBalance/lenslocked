package main

import (
	"fmt"

	"github.com/lifebalance/lenslocked/models"
)

func main() {
	gs := models.GalleryService{}
	imgs, err := gs.Images(2)
	if err != nil {
		panic(err)
	}
	for _, i := range imgs {
		fmt.Println(i.Path, "\t\t", i.Filename)
	}
}
