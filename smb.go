package main

import (
	"fmt"
	"log"
	"mortar/clients"
)

func test() {
	filters := []string{".DS_Store"}

	client, err := clients.NewSMBClient("192.168.1.20", 445,
		"GUEST", "", "guest", filters)

	defer client.Close()
	if err != nil {
		log.Fatal(err)
	}

	ls, err := client.ListDirectory("GBA")
	if err != nil {
		fmt.Println(err)
	}

	for _, l := range ls {
		fmt.Printf("%s\n", l.Filename)
	}

	err = client.DownloadFile("GBA/Mother 3.gba", "ROMS/", "Mother 3.gba")
	if err != nil {
		fmt.Println(err)
	}
}
