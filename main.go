package main

import (
	"log"
	"techvision/balancer/global"
	"techvision/balancer/parse"

	_ "github.com/joho/godotenv/autoload"
)

var GlobalNodes global.Nodes = parse.ParseNodes()

func main() {
	log.Println("started")
	SetupServer()
}
