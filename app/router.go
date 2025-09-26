package app

import (
	"net/http"
	"os"
	"synk/gateway/app/controller"
	"synk/gateway/app/util"
)

func Router(service *Service) {
	aboutController := controller.NewAbout(service.DB)
	postController := controller.NewPosts(service.DB)

	http.HandleFunc("GET /about", aboutController.HandleAbout)
	http.HandleFunc("GET /post", postController.HandleList)

	util.Log("app running on port 8080 to " + os.Getenv("PORT"))

	http.ListenAndServe(":8080", nil)
}
