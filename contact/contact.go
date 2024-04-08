package contact

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const PageSize = 100

type Contact struct {
	ID     int    `json:"id"`
	First  string `json:"first"`
	Last   string `json:"last"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Errors map[string]string
}

var db = make(map[int]*Contact)

func (c *Contact) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *Contact) Update(first, last, phone, email string) {
	c.First = first
	c.Last = last
	c.Phone = phone
	c.Email = email
}

func (c *Contact) Validate() bool {
	c.Errors = make(map[string]string)
	if c.Email == "" {
		c.Errors["Email"] = "Email Required"
	}
	for _, contact := range db {
		if contact.ID != c.ID && contact.Email == c.Email {
			c.Errors["Email"] = "Email Must Be Unique"
			break
		}
	}
	return len(c.Errors) == 0
}

func (c *Contact) Save() bool {
	if !c.Validate() {
		return false
	}
	if c.ID == 0 {
		maxID := 0
		for id := range db {
			if id > maxID {
				maxID = id
			}
		}
		c.ID = maxID + 1
	}
	db[c.ID] = c
	saveDB()
	return true
}

func (c *Contact) Delete() {
	delete(db, c.ID)
	saveDB()
}

func Count() int {
	time.Sleep(2 * time.Second)
	return len(db)
}

func All(page int) []*Contact {
	start := (page - 1) * PageSize
	end := start + PageSize
	if start >= len(db) {
		return nil
	}
	if end > len(db) {
		end = len(db)
	}
	var contacts []*Contact
	for _, c := range db {
		contacts = append(contacts, c)
	}
	return contacts[start:end]
}

func Search(text string) []*Contact {
	var result []*Contact
	for _, c := range db {
		if (c.First != "" && strings.Contains(c.First, text)) ||
			(c.Last != "" && strings.Contains(c.Last, text)) ||
			(c.Email != "" && strings.Contains(c.Email, text)) ||
			(c.Phone != "" && strings.Contains(c.Phone, text)) {
			result = append(result, c)
		}
	}
	return result
}

func LoadDB() {
	// Open the contacts.json file for reading
	file, err := os.Open("contacts.json")
	if err != nil {
		fmt.Println("Error opening contacts.json:", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into a slice of Contact structs
	var contacts []*Contact
	err = json.NewDecoder(file).Decode(&contacts)
	if err != nil {
		fmt.Println("Error decoding contacts.json:", err)
		return
	}

	// Populate the db map with the loaded contacts
	for _, contact := range contacts {
		db[contact.ID] = contact
	}
}

func saveDB() {
	// Open the contacts.json file for writing
	file, err := os.OpenFile("contacts.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening contacts.json:", err)
		return
	}
	defer file.Close()

	// Encode the db map into JSON and write it to the file
	var contacts []*Contact
	for _, contact := range db {
		contacts = append(contacts, contact)
	}
	err = json.NewEncoder(file).Encode(contacts)
	if err != nil {
		fmt.Println("Error encoding contacts.json:", err)
		return
	}
}

func FindByID(id int) *Contact {
	contact, ok := db[id]
	if !ok {
		return nil
	}
	contact.Errors = make(map[string]string)
	return contact
}
