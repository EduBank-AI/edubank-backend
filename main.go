package main

import (
	"fmt"
	"log"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("image.png")

	text, err := client.Text()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(text)
}
