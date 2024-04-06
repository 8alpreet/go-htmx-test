package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	contact "github.com/8alpreet/go-htmx-test/contact"
)

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contacts", http.StatusFound)
}

func showContacts(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("showContacts handler called")

	data := struct {
		Query    string
		Contacts []*contact.Contact
	}{
		Query:    r.URL.Query().Get("q"),
		Contacts: contact.All(1),
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
	err := tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Fatal(err)
	}

}

func showContactByID(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("showContactByID handler called")

	idStr := r.URL.Path[len("/contacts/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	contact := contact.FindByID(id)
	if contact == nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/show.html"))
	err = tmpl.ExecuteTemplate(w, "layout.html", contact)
	if err != nil {
		log.Fatal(err)
	}

}

func newContact(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("newContact handler called. method: ", r.Method)

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/new.html"))
	if r.Method == http.MethodGet {
		err := tmpl.ExecuteTemplate(w, "layout.html", contact.Contact{})
		if err != nil {
			log.Fatal(err)
		}
	}

	if r.Method == http.MethodPost {
		newC := contact.Contact{
			First: r.FormValue("firstName"),
			Last:  r.FormValue("lastName"),
			Phone: r.FormValue("phone"),
			Email: r.FormValue("email"),
		}
		if ok := newC.Save(); ok {
			log.Default().Println("Contact saved successfully")
			http.Redirect(w, r, "/contacts", http.StatusFound)
		} else {
			err := tmpl.ExecuteTemplate(w, "layout.html", contact.Contact{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func main() {
	contact.LoadDB()
	http.HandleFunc("/", index)
	http.HandleFunc("/contacts", showContacts)
	http.HandleFunc("/contacts/new", newContact)
	http.HandleFunc("/contacts/{id}", showContactByID)

	fmt.Println("Server is running on: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
