package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

const webPort = "8080"

var hashAndURL = make(map[string]string)

func main() {

	// Since net/http or HandleFunc can't handle dynamic value or dynamic url like {urlShortCode}, we're left
	// with no other option but to use a router package. Here, I'm using go-chi router
	// http.HandleFunc("/{urlShortCode}", RedirectIfURLFound)
	// http.HandleFunc("/getShortUrl", GetShortURLHandler)

	fmt.Printf("starting webserver on htttp://localhost:%s ...\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Fatalf("could not start the http server %v\n", err)
	}

}

func routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Post("/getShortUrl", GetShortURLHandler)
	mux.Get("/{urlShortCode}", RedirectIfURLFound)

	return mux
}

func GetShortURLHandler(w http.ResponseWriter, r *http.Request) {
	type URLRequestObject struct {
		URL string `json:"url"`
	}

	type URLCollection struct {
		ActualURL string
		ShortURL  string
	}

	type SuccessResponse struct {
		Code     int
		Message  string
		Response URLCollection
	}

	var urlRequest URLRequestObject
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
		http.Error(w, "Posted data not supported", http.StatusFound)
		return
	}
	err = json.Unmarshal(body, &urlRequest)
	if err != nil {
		// log.Println("The Request Body:", string(body))
		log.Panic("error occured while unmarshalling in GetShortURLHandler() :", err)

		http.Error(w, "Posted URL not supported", http.StatusNotFound) // "URL can't be empty"
		return
	}

	if !(isURL(urlRequest.URL)) {
		http.Error(w, "An invalid URL found, provide a valid URL", http.StatusNotFound)
		return
	}

	uniqueID := uuid.New().String()[:8] // shortURL or shortcode
	hashAndURL[uniqueID] = urlRequest.URL

	successResponse := SuccessResponse{
		Code:    http.StatusAccepted,
		Message: "short url generated for the provided URL",
		Response: URLCollection{
			ActualURL: urlRequest.URL,
			ShortURL:  r.Host + "/" + uniqueID,
		},
	}

	jsonResponse, _ := json.Marshal(successResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(successResponse.Code)
	w.Write(jsonResponse)

}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func RedirectIfURLFound(w http.ResponseWriter, r *http.Request) {

	// fmt.Println("url query:", r.URL.Query())
	// shortURL := r.URL.Query().Get("urlShortCode")
	// url, ok := hashAndURL[shortURL]
	// if ok {
	// 	http.Redirect(w, r, url, http.StatusSeeOther)
	// } else {
	// 	http.Error(w, "Not found", http.StatusNotFound)
	// }

	// fmt.Println("request:", r)
	fmt.Println("url path:", r.URL.Path)
	path := r.URL.Path
	shortURL := strings.TrimPrefix(path, "/")
	url, ok := hashAndURL[shortURL]
	if ok {
		http.Redirect(w, r, url, http.StatusSeeOther)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
	}

}
