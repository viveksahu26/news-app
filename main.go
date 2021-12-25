package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/freshman-tech/news-demo-starter-files/news"
	"github.com/joho/godotenv"
)

type Search struct {
	Query      string
	NextPage   int
	TotalPages int
	Results    *news.Results
}

var newsapi *news.Client

// tpl is a package level variable that points to a template definition from the provided files

//template.ParseFiles is wrapped with template.Must so that the code panics if an error is obtained while parsing the template file
var tpl = template.Must(template.ParseFiles("index.html"))

// To determine if the last page of results
func (s *Search) IsLastPage() bool {
	return s.NextPage >= s.TotalPages
}

func (s *Search) CurrentPage() int {
	if s.NextPage == 1 {
		return s.NextPage
	}

	return s.NextPage - 1
}

func (s *Search) PreviousPage() int {
	return s.CurrentPage() - 1
}

func searchHandler(newsapi *news.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse((r.URL.String()))
		fmt.Println("r.URL.String(): ", r.URL.String())
		fmt.Println("u ----> ", u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params := u.Query()
		fmt.Println("param: ", params)
		searchQuery := params.Get("q")
		page := params.Get("page")
		if page == "" {
			page = "1"
		}

		fmt.Println("Search Query is: ", searchQuery)
		fmt.Println("Page is: ", page)

		results, err := newsapi.FetchEverything(searchQuery, page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		nextPage, err := strconv.Atoi(page)
		fmt.Println("nextPage: ", nextPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		search := &Search{
			Query:      searchQuery,
			NextPage:   nextPage,
			TotalPages: int(math.Ceil(float64(results.TotalResults) / float64(newsapi.PageSize))),
			Results:    results,
		}

		if ok := !search.IsLastPage(); ok {
			search.NextPage++
		}

		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
	}
}

// w --> it send response to the client request
// r --> it accept request from client
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("<h1>Hi vivek!! This is your first web page..!</h1>"))
	// w.Write([]byte("<h2>Change in the Port nubmer..</h2>"))
	// w.Write([]byte("News App API..."))

	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
	// two arguments: where we want to write the output to, and the data we want to pass to the template.
	// tpl.Execute(w, nil)
}

func main() {
	// Load() --> reads the .env file and loads the set variables into the environment
	err := godotenv.Load()
	if err != nil {
		log.Println("Error in loading .env files..")
	}
	fmt.Println("PORT is: ", os.Getenv("PORT"))
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiKey := os.Getenv("NEWS_API_KEY")
	if apiKey == "" {
		log.Fatal("Env: apiKey must be set")
	}

	myClient := &http.Client{Timeout: 10 * time.Second}
	newsapi = news.NewClient(myClient, apiKey, 20)

	fs := http.FileServer(http.Dir("assets"))
	// create an HTTP request multiplexer which is subsequently assigned to the mux variable
	// calls the associated handler for the pattern whenever a match is found.
	mux := http.NewServeMux()

	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/search", searchHandler(newsapi))
	// mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)

	// starts the server on the port
	http.ListenAndServe(":"+port, mux)
}
