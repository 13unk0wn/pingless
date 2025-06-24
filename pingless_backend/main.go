package main

import (
	"log"
	"pingless/config"
	"pingless/db"
	"pingless/routes"
)

func main() {
	db, err := db.Init()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(db)
	log.Println("DB SETUP SUCCESSFUL")
	config := config.LoadConfig(db)
	log.Println(config)

	routes.Routes(db)
}
