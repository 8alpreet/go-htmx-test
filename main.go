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

	search := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 1
	}
	nextPage := page + 1

	gotContacts := []*contact.Contact{}
	if search == "" {
		gotContacts = contact.All(page)
	} else {
		gotContacts = contact.Search(search)
	}
	
	// need to check if we they are furtehr pages of contacts to display
	totalPages := contact.Count() / contact.PageSize
	if contact.Count() % contact.PageSize > 0 {
		totalPages++
	}

	data := struct {
		PrevPage   int
		Page       int
		NextPage   int
		TotalPages int
		Query      string
		Contacts   []*contact.Contact
	}{
		PrevPage:   prevPage,
		Page:       page,
		NextPage:   nextPage,
		TotalPages: totalPages,
		Query:      search,
		Contacts:   gotContacts,
	}

	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html"))
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
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
			log.Default().Println("Error saving contact. Errors: ", newC.Errors)
			err := tmpl.ExecuteTemplate(w, "layout.html", newC)
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
			log.Default().Println("Error saving contact. Errors: ", c.Errors)
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
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	idStr := strings.TrimPrefix(r.URL.Path, "/contacts/")
	idStr = strings.TrimSuffix(idStr, "/delete")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	contact.FindByID(id).Delete()
	http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}

func getContactEmail(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("getContactsEmail handler called")

	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/contacts/")
	idStr = strings.TrimSuffix(idStr, "/email")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	c := contact.FindByID(id)
	if c == nil {
		c = &contact.Contact{}
	}
	c.Email = r.URL.Query().Get("email")
	c.Validate()

	fmt.Fprintf(w, "%s", c.Errors["Email"])
}

func sharedContactRequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		showContactByID(w, r)
	case http.MethodDelete:
		deleteContact(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	contact.LoadDB()
	http.HandleFunc("/", index)
	http.HandleFunc("/contacts", showContacts)
	http.HandleFunc("/contacts/new", newContact)        // GET or POST
	http.HandleFunc("/contacts/{id}/edit", editContact) // GET or POST
	http.HandleFunc("/contacts/{id}", sharedContactRequestHandler)
	http.HandleFunc("/contacts/{id}/email", getContactEmail)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server is running on: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
