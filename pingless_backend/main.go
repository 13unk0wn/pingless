package main

import (
	"log"
	"pingless/config"
	"pingless/db"
)

func main() {
	config := config.LoadConfig()
	log.Println(config)
	db, err := db.Init()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(db)
	log.Println("DB SETUP SUCCESSFUL")
}
