package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.New("t").Funcs(template.FuncMap{"formatBody": func(body []byte) string {
	for _, instance := range availablePages {
		// formattedBody
		body = []byte(strings.ReplaceAll(string(body), instance, "<a href=\"/view/"+instance+"\">"+instance+"</a>"))
	}
	return string(body)
}}).ParseFiles("templates/edit.html", "templates/view.html"))
var validPath = regexp.MustCompile("^/(edit|view|save)/([a-zA-Z0-9]+)$")
var availablePages []string

// store names of available pages, find instances of names in page body, replace instances with link to page

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func load(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, error := os.ReadFile(filename)
	if error != nil {
		return nil, error
	}

	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := load(title)
	// for _, instance := range availablePages {
	// 	// formattedBody
	// 	page.Body = []byte(strings.ReplaceAll(string(page.Body), instance, "<a href=\"/view/"+instance+"\">"+instance+"</a>"))
	// }
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}
	renderTemplate(w, "view", page)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	page, err := load(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(w, "edit", page)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, templateName string, page *Page) {
	err := templates.ExecuteTemplate(w, templateName+".html", page)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(w http.ResponseWriter, r *http.Request, title string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
		}
		fn(w, r, m[2])
	}
}

func formatBody(body []byte) string {
	for _, instance := range availablePages {
		// formattedBody
		body = []byte(strings.ReplaceAll(string(body), instance, "<a href=\"/view/"+instance+"\">"+instance+"</a>"))
	}
	return string(body)
}

func main() {
	var pagesDir, err = os.Open("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer pagesDir.Close()
	fileInfo, err := pagesDir.ReadDir(-1)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range fileInfo {
		availablePages = append(availablePages, strings.TrimSuffix(file.Name(), ".txt"))
	}

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
