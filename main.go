package main

import (
	"github.com/diogovalentte/dashboard/api"
)

func main() {
	router := api.SetupRouter()

	router.Run()
}
