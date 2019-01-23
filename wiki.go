package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".md"
	return ioutil.WriteFile("pages/"+filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".md"
	body, err := ioutil.ReadFile("pages/" + filename)
	if err != nil {
		log.Println("Error loading file")
		return nil, err
	}
	content := bluemonday.UGCPolicy().SanitizeBytes(body)
	html := blackfriday.Run(content)
	return &Page{Title: title, Body: html}, nil
}

func loadPageContent(title string) ([]byte, error) {
	filename := title + ".md"
	body, err := ioutil.ReadFile("pages/" + filename)
	if err != nil {
		log.Println("Error loading file")
		return nil, err
	}

	return body, nil
}

func loadPages() ([]*Page, error) {
	pages := []*Page{}

	files, err := ioutil.ReadDir("pages")
	if err != nil {
		log.Printf("Error loading pages: %s\n", err)
	}

	for _, file := range files {
		title := file.Name()[:strings.Index(file.Name(), ".")]

		page, err := loadPage(title)
		if err != nil {
			log.Printf("Error loading page %s: %s", title, err)
		}

		pages = append(pages, page)
	}

	return pages, nil
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	pages, err := loadPages()

	if err != nil {
		log.Printf("Could not load pages, %s\n", err)
	}

	renderPages(w, "home", pages)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]

	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}

	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	content, err := loadPageContent(title)

	if err != nil {
		content = []byte{}
		log.Println(err)
	}

	page := &Page{Title: title, Body: content}

	renderTemplate(w, "edit", page)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles("templates/layout.html", "templates/"+tmpl+".html")

	if err != nil {
		log.Println(err)
	}

	t.ExecuteTemplate(w, "layout", p)
}

func renderPages(w http.ResponseWriter, tmpl string, pages []*Page) {
	t, err := template.ParseFiles("templates/layout.html", "templates/"+tmpl+".html")

	if err != nil {
		log.Println(err)
	}

	t.ExecuteTemplate(w, "layout", pages)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
