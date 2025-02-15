package web

import (
	"context"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vituchon/splitify/presentation/web/controllers"
	"github.com/vituchon/splitify/util"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

const (
	storeKeyFilePath = ".ss" // the file were the actual key is stored
)

func retrieveCookieStoreKey(filepath string) (key []byte, err error) {
	if util.FileExists(filepath) {
		key, err = ioutil.ReadFile(storeKeyFilePath)
	} else {
		key = securecookie.GenerateRandomKey(32)
		ioutil.WriteFile(storeKeyFilePath, key, 0644)
	}
	return
}

func StartServer() {
	key, err := retrieveCookieStoreKey(storeKeyFilePath)
	if err != nil {
		log.Printf("Unexpected error while retrieving cookie store key: %v", err)
		return
	}
	controllers.InitSessionStore(key)

	router := buildRouter()
	port := getenv("PORT", "9999")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  40 * time.Second,
		WriteTimeout: 300 * time.Second,
	}
	log.Printf("Splitify web server listening at port %v", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		log.Println("Unexpected error initiliazing piedra papel y tijera web server: ", err)
	}

	// TODO (for greater good) : Perhaps we are now in condition to add https://github.com/gorilla/mux#graceful-shutdown
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func buildRouter() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(NoMatchingHandler)

	assetsFileServer := http.FileServer(http.Dir("./presentation/web/assets"))
	assetsRouter := router.PathPrefix("/assets").Subrouter()
	assetsRouter.PathPrefix("/").Handler(http.StripPrefix("/assets", assetsFileServer))

	rootRouter := router.PathPrefix("/").Subrouter()
	rootRouter.Use(AccessLogMiddleware, ClientSessionAwareMiddleware)

	rootGet := BuildSetHandleFunc(rootRouter, "GET")
	rootGet("/", serveRoot)
	rootGet("/healthcheck", controllers.Healthcheck)
	rootGet("/version", controllers.Version)

	apiRouter := rootRouter.PathPrefix("/api/v1").Subrouter()
	apiGet := BuildSetHandleFunc(apiRouter, "GET")
	apiPost := BuildSetHandleFunc(apiRouter, "POST")
	//apiDelete := BuildSetHandleFunc(apiRouter, "DELETE")

	apiGet("/groups", controllers.GetAllGroups)
	apiPost("/groups", controllers.CreateGroup)
	apiGet("/groups/{groupId:[0-9]+}/participants", controllers.GetGroupParticipants)
	apiPost("/groups/{groupId:[0-9]+}/participants", controllers.AddParcipantToGroup)
	return router
}

type setHandlerFunc func(path string, f http.HandlerFunc)

// Creates a function for register a handler for a path for the given router and http methods
func BuildSetHandleFunc(router *mux.Router, methods ...string) setHandlerFunc {
	return func(path string, f http.HandlerFunc) {
		router.HandleFunc(path, f).Methods(methods...)
	}
}

func NoMatchingHandler(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/favicon.ico" { // don't log this
		log.Println("No maching route for " + request.URL.Path)
	}
	response.WriteHeader(http.StatusNotFound)
}

// Adds a logging handler for logging each request's in Apache Common Log Format (CLF).
// With this middleware we ensure that each requests will be, at least, logged once.
func AccessLogMiddleware(h http.Handler) http.Handler {
	loggingHandler := handlers.LoggingHandler(os.Stdout, h)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggingHandler.ServeHTTP(w, r)
	})
}

func ClientSessionAwareMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		clientSession, err := controllers.GetOrCreateClientSession(request)
		if err != nil {
			log.Printf("error while getting client session: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = controllers.SaveClientSession(request, response, clientSession)
		if err != nil {
			log.Printf("error while saving client session: %v", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(request.Context(), "clientSession", clientSession)
		h.ServeHTTP(response, request.WithContext(ctx))
	})
}

func serveRoot(response http.ResponseWriter, request *http.Request) {
	t, err := template.ParseFiles("./presentation/web/assets/index.html")
	if err != nil {
		log.Printf("Error while parsing template : %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.Execute(response, nil)
}
