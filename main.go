package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	contact "github.com/8alpreet/go-htmx-test/contact"
)

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/contacts", http.StatusFound)
}

func add(a, b int) int {
	return a + b
}
func mul(a float64, b int) int {
	return int(a * float64(b))
}

func renderHomeTemplate(w http.ResponseWriter, data interface{}) {
	funcs := template.FuncMap{
		"add": add,
		"mul": mul,
	}
	tmpl := template.New("layout.html").Funcs(funcs)
	tmpl = template.Must(tmpl.ParseFiles("templates/layout.html", "templates/index.html", "templates/rows.html", "templates/archiveUI.html"))
	err := tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Fatal(err)
	}
}

func showContacts(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("showContacts handler called")

	search := r.URL.Query().Get("q")
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	data := struct {
		Page     int
		PageSize int
		Query    string
		Contacts []*contact.Contact
		Archiver *contact.Archiver
	}{
		Page:     page,
		PageSize: contact.PageSize,
		Query:    search,
		Archiver: contact.NewArchiver(),
	}
	if search == "" {
		data.Contacts = contact.All(page)
	} else {
		data.Contacts = contact.Search(search)
	}

	// when there is a search, we don't need to return the whole page
	// just the table rows
	// so the authors nest this check for HX-Trigger under thier
	// check for Search != None is python. This causes trouble in Go
	// becuase the search comes as an empty "" when the user erases
	// the field, which causes the whole page to be rendered within the
	// table
	if r.Header.Get("HX-Trigger") == "search" {
		tmpl := template.Must(template.ParseFiles("templates/rows.html"))
		err := tmpl.ExecuteTemplate(w, "rows", data)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// provides custom funcs for the template
	// this was used to provide the "add" function to the template
	// funcs := template.FuncMap{
	// 	"add": add,
	// }
	// Returning the full template
	renderHomeTemplate(w, data)
	// tmpl := template.New("layout.html").Funcs(funcs)
	// tmpl = template.Must(tmpl.ParseFiles("templates/layout.html", "templates/index.html", "templates/rows.html"))
	// err = tmpl.ExecuteTemplate(w, "layout.html", data)
	// if err != nil {
	// 	log.Fatal(err)
	// }

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

	if r.Method != http.MethodDelete {
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

	// request is coming for delete button in edit.html
	// so we redirect
	if r.Header.Get("HX-Trigger") == "delete-btn" {
		http.Redirect(w, r, "/contacts", http.StatusSeeOther)
		return
	}
	// request is coming from inline delete button in rows.html
	// so we don't need to redirect
	fmt.Fprint(w, "")
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

func deleteSelectedContacts(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("deleteSelectedContacts handler called")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Handle error
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Convert the body to a string
	bodyStr := string(body)

	// Parse the URL-encoded string
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusInternalServerError)
		return
	}

	// Extract the selected_contact_ids
	selectedContactIds := values["selected_contact_ids"]
	for _, id := range selectedContactIds {
		intID, err := strconv.Atoi(id)
		if err != nil {
			log.Default().Printf("Error converting ID %s to int: %v", id, err)
			continue
		}
		contact.FindByID(intID).Delete()
	}

	data := map[string]interface{}{
		"Page":     1,
		"PageSize": contact.PageSize,
		"Query":    "",
		"Contacts": contact.All(1),
	}
	tmpl := template.Must(template.ParseFiles("templates/layout.html", "templates/index.html", "templates/rows.html"))
	err = tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Fatal(err)
	}
}

func sharedContactIdHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		showContactByID(w, r)
	case http.MethodDelete:
		deleteContact(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func sharedContactHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("sharedContactHandler handler called")

	switch r.Method {
	case http.MethodDelete:
		deleteSelectedContacts(w, r)
	case http.MethodGet:
		showContacts(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

func getContactCount(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("getContactCount handler called")
	time.Sleep(2 * time.Second) // simulate a slow network call
	fmt.Fprintf(w, "( %d total Contacts )", contact.Count())
}

func sharedArchiveHandler(w http.ResponseWriter, r *http.Request) {
	log.Default().Println("sharedArchiveHandler called")

	funcs := template.FuncMap{
		"mul": mul,
	}
	tmpl := template.New("archiveUI.html").Funcs(funcs)
	tmpl = template.Must(tmpl.ParseFiles("templates/archiveUI.html"))
	if r.Method == http.MethodPost {
		arch := contact.NewArchiver()
		arch.Run()
		data := map[string]interface{}{
			"Archiver": arch,
		}

		err := tmpl.ExecuteTemplate(w, "archive-ui", data)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.Method == http.MethodGet {
		arch := contact.NewArchiver()
		data := map[string]interface{}{
			"Archiver": arch,
		}
		err := tmpl.ExecuteTemplate(w, "archive-ui", data)
		if err != nil {
			log.Fatal(err)
		}
		return

	}

	if r.Method == http.MethodDelete {
		arch := contact.NewArchiver()
		arch.Reset()
		data := map[string]interface{}{
			"Archiver": arch,
		}
		err := tmpl.ExecuteTemplate(w, "archive-ui", data)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}
// the file opens in the browser instead of a typical download
// but I think that's acceptable for now
func downloadArchiveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Default().Println("downloadArchiveFile handler called")
	http.ServeFile(w, r, "contacts.json")
}

func main() {
	contact.LoadDB()
	http.HandleFunc("/", index)
	http.HandleFunc("/contacts", sharedContactHandler)
	http.HandleFunc("/contacts/new", newContact)        // GET or POST
	http.HandleFunc("/contacts/{id}/edit", editContact) // GET or POST
	http.HandleFunc("/contacts/{id}", sharedContactIdHandler)
	http.HandleFunc("/contacts/{id}/email", getContactEmail)
	http.HandleFunc("/contacts/count", getContactCount)
	http.HandleFunc("/contacts/archive", sharedArchiveHandler)     // POST or GET
	http.HandleFunc("/contacts/archive/file", downloadArchiveFile) // GET
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server is running on: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
