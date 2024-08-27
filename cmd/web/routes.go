package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)
	// unprotected routes
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.handleHome))
	router.Handler(http.MethodGet, "/snippet/view/:id", dynamic.ThenFunc(app.snippetView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.signUpUser))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.signUpUserPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.logInUser))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.logInUserPost))

	protected := dynamic.Append(app.requireAuth)
	// protected routes
	router.Handler(http.MethodGet, "/snippet/create", protected.ThenFunc(app.snippetCreate))
	router.Handler(http.MethodPost, "/snippet/create", protected.ThenFunc(app.snippetCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.logoutUserPost))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
