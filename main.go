package main

import (
	"net/http"

	"github.com/imba3r/pkg/scheduler"
)

func main() {
	schedulerService := scheduler.NewService()

	http.ListenAndServe(":1234", schedulerService.Handler())
}
