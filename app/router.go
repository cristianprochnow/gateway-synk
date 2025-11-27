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
	intCredentialController := controller.NewIntCredentials(service.DB)

	http.HandleFunc("GET /about", aboutController.HandleAbout)
	http.HandleFunc("GET /post", postController.HandleList)
	http.HandleFunc("POST /post", postController.HandleCreate)
	http.HandleFunc("PUT /post", postController.HandleUpdate)
	http.HandleFunc("DELETE /post", postController.HandleDelete)
	http.HandleFunc("POST /post/publish", postController.HandlePublish)
	http.HandleFunc("GET /templates/basic", templateController.HandleBasicList)
	http.HandleFunc("GET /templates", templateController.HandleList)
	http.HandleFunc("POST /templates", templateController.HandleCreate)
	http.HandleFunc("PUT /templates", templateController.HandleUpdate)
	http.HandleFunc("DELETE /templates", templateController.HandleDelete)
	http.HandleFunc("GET /int_profiles/basic", intProfileController.HandleBasicList)
	http.HandleFunc("GET /int_profiles", intProfileController.HandleList)
	http.HandleFunc("POST /int_profiles", intProfileController.HandleCreate)
	http.HandleFunc("PUT /int_profiles", intProfileController.HandleUpdate)
	http.HandleFunc("DELETE /int_profiles", intProfileController.HandleDelete)
	http.HandleFunc("GET /int_credentials/basic", intCredentialController.HandleBasicList)
	http.HandleFunc("GET /int_credentials", intCredentialController.HandleList)
	http.HandleFunc("POST /int_credentials", intCredentialController.HandleCreate)
	http.HandleFunc("PUT /int_credentials", intCredentialController.HandleUpdate)
	http.HandleFunc("DELETE /int_credentials", intCredentialController.HandleDelete)

	port := os.Getenv("PORT")
	util.Log("app running on port " + port)

	err := http.ListenAndServeTLS(
		":"+port,
		"/cert/cert.pem",
		"/cert/key.pem",
		controller.Cors(http.DefaultServeMux),
	)
	if err != nil {
		util.Log("app failed on running on port " + port + ": " + err.Error())
	}
}
