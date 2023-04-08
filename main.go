package main

import (
	"github.com/joho/godotenv"
	"github.com/joshDamian/web3storage-service/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
	app.App().Run(":8080")
}
