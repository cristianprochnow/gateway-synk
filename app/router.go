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
	templateController := controller.NewTemplates(service.DB)
	intProfileController := controller.NewIntProfiles(service.DB)

	http.HandleFunc("GET /about", aboutController.HandleAbout)
	http.HandleFunc("GET /post", postController.HandleList)
	http.HandleFunc("POST /post", postController.HandleCreate)
	http.HandleFunc("PUT /post", postController.HandleUpdate)
	http.HandleFunc("DELETE /post", postController.HandleDelete)
	http.HandleFunc("GET /templates/basic", templateController.HandleBasicList)
	http.HandleFunc("GET /templates", templateController.HandleList)
	http.HandleFunc("POST /templates", templateController.HandleCreate)
	http.HandleFunc("PUT /templates", templateController.HandleUpdate)
	http.HandleFunc("GET /int_profiles/basic", intProfileController.HandleBasicList)

	util.Log("app running on port 8080 to " + os.Getenv("PORT"))

	err := http.ListenAndServe(":8080", controller.Cors(http.DefaultServeMux))
	if err != nil {
		util.Log("app failed on running on port 8080: " + err.Error())
	}
}
