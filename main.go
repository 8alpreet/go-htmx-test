package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

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
		return
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
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func editContact(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("editContact handler called")

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/edit.html"))
	idStr := strings.TrimPrefix(r.URL.Path, "/contacts/")
	idStr = strings.TrimSuffix(idStr, "/edit")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	c := contact.FindByID(id)
	if c == nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}
	if r.Method == http.MethodGet {
		err := tmpl.ExecuteTemplate(w, "layout.html", c)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		// r.ParseForm()
		// for key, values := range r.Form {
		// 	for _, value := range values {
		// 		log.Default().Printf("%s: %s\n", key, value)
		// 	}
		// }
		c.Update(r.FormValue("first"), r.FormValue("last"), r.FormValue("phone"), r.FormValue("email"))
		if !c.Save() {
			log.Default().Println("Error updating contact. Errors: ", c.Errors)
			err := tmpl.ExecuteTemplate(w, "layout.html", c)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		path := fmt.Sprintf("/contacts/%d", c.ID)
		http.Redirect(w, r, path, http.StatusFound)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

func deleteContact(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("deleteContact handler called")
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/contacts/")
	idStr = strings.TrimSuffix(idStr, "/delete")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	contact.FindByID(id).Delete()
	http.Redirect(w, r, "/contacts", http.StatusFound)
}

func main() {
	contact.LoadDB()
	http.HandleFunc("/", index)
	http.HandleFunc("/contacts", showContacts)
	http.HandleFunc("/contacts/new", newContact) // GET or POST
	http.HandleFunc("/contacts/{id}", showContactByID)
	http.HandleFunc("/contacts/{id}/edit", editContact)     // GET or POST
	http.HandleFunc("/contacts/{id}/delete", deleteContact) // POST

	fmt.Println("Server is running on: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
