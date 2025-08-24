package app

import (
	"net/http"
	"os"
	"synk/gateway/app/controller"
	"synk/gateway/app/util"
)

func Router(service *Service) {
	aboutController := controller.NewAbout(service.DB)

	http.HandleFunc("GET /about", aboutController.HandleAbout)

	util.Log("app running on port 8080 to " + os.Getenv("PORT"))

	http.ListenAndServe(":8080", nil)
}
