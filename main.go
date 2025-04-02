package main

import (
	"log"
	"techvision/balancer/global"
	"techvision/balancer/parse"
	_ "techvision/balancer/sync"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	global.GNodes = parse.ParseNodes()
	log.Println("started")
	SetupServer()
}
