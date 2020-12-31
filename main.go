package main

import (
	"fmt"
	"qbt/qbt"
)

func main() {
	client, err := qbt.New("http://127.0.0.1", "admin", "admin")
	if err != nil {
		fmt.Println(err)
	} else {
		list, _ := client.List()
		fmt.Println(list)
	}
}
