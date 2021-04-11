package api

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sort"
)

var errAlreadyExists = errors.New("Item already exists")
var errMissingField = errors.New("Item is missing a required field")
var errDoesNotExist = errors.New("Item does not exist")

// Book type with Name, Author and ISBN
type Book struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	ISBN   string `json:"isbn"`
	// define book
}

// Books slice of all known books
var Books = map[string]Book{
	"0123456789": Book{Title: "Cloud Native Go", Author: "M. L. Reimer", ISBN: "0123456789"},
	"0987654321": Book{Title: "Hello World", Author: "E. Pavlova", ISBN: "0987654321"},
}

// ToJSON to be used for marshalling of Book Type
func (b Book) ToJSON() []byte {
	json, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}
	return json
}

// FromJSON to be used for unmarshalling of Book type
func FromJSON(data []byte) Book {
	book := Book{}
	err := json.Unmarshal(data, &book)
	if err != nil {
		panic(err)
	}
	return book
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// BooksHandleFunc to be used as http.HandleFunc for Book API (Get all books, or add a book)
func BooksHandleFunc(w http.ResponseWriter, r *http.Request) {
	setupCORS(&w, r)

	switch method := r.Method; method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
		return
	// GET all books
	case http.MethodGet:
		books := AllBooks()
		b, err := json.Marshal(books)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		w.Header().Add("Content-Type", "application/json; charset-utf-8")
		w.Write(b)
	// POST a new book to the list of books
	case http.MethodPost:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		book := FromJSON(body)
		b, err := CreateBook(book)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			if err == errAlreadyExists {
				w.Write([]byte("Item already exists"))
			} else if err == errMissingField {
				w.Write([]byte("Item is missing a required field"))
			} else {
				w.Write([]byte("There was an error creating the item"))
			}
			break
		}
		jsonB := b.ToJSON()
		// send back the created book
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json; charset-utf-8")
		w.Write(jsonB)

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported request method."))
	}
}

// BookHandleFunc to be used as http.HandleFunc for Book API (Get a single book, or edit a book, or delete a book)
func BookHandleFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// extract ISBN as the last part of the url
	isbn := r.URL.Path[len("/api/books/"):]
	// switch over request method
	switch method := r.Method; method {
	// GET a single book
	case http.MethodGet:
		b, found := GetBook(isbn)
		if found {
			jsonB := b.ToJSON()
			// send back the located book
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json; charset-utf-8")
			w.Write(jsonB)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Item was not found."))
		}
	// PUT update a book
	case http.MethodPut:
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			break
		}
		book := FromJSON(body)
		b, err := UpdateBook(isbn, book)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			if err == errDoesNotExist {
				w.Write([]byte("Item does not exist"))
			} else if err == errMissingField {
				w.Write([]byte("Item is missing a required field"))
			} else {
				w.Write([]byte("There was an error modifying the item"))
			}
			break
		}
		jsonB := b.ToJSON()
		// send back the updated book
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json; charset-utf-8")
		w.Write(jsonB)
	// DELETE a book from the Books
	case http.MethodDelete:
		b, err := DeleteBook(isbn)
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			if err == errDoesNotExist {
				w.Write([]byte("Item does not exist"))
			} else {
				w.Write([]byte("There was an error deleting the item"))
			}
			break
		}
		jsonB := b.ToJSON()
		// send back the updated book
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json; charset-utf-8")
		w.Write(jsonB)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unsupported request method."))
	}
}

// AllBooks returns a slice of all books
func AllBooks() []Book {
	// initialize an empty slice of Books with the length equal to the length of the Books map
	s := make([]Book, len(Books))
	idx := 0
	// loop over the Books map and enter each book into the slice
	for _, value := range Books {
		s[idx] = value
		idx++
	}
	sort.Slice(s, func(i, j int) bool {
		return s[i].Title < s[j].Title
	})
	return s
}

// CreateBook to be used for adding a book to the books dictionary
func CreateBook(b Book) (Book, error) {
	// check to make sure that user provided author, isbn and title
	if len(b.Author) < 1 || len(b.ISBN) < 1 || len(b.Title) < 1 {
		return Book{}, errMissingField
	}
	isbn := b.ISBN
	// check if the book with the provided ISBN already exists
	_, ok := Books[isbn]
	// if it already exists, throw error
	if ok == true {
		return Book{}, errAlreadyExists
	}
	Books[isbn] = b
	return b, nil
}

// GetBook returns the book for a given ISBN
func GetBook(isbn string) (Book, bool) {
	b, ok := Books[isbn]
	return b, ok
}

// UpdateBook will update a book
func UpdateBook(isbn string, b Book) (Book, error) {
	// check to make sure that user provided author, isbn and title
	if len(b.Author) < 1 || len(b.ISBN) < 1 || len(b.Title) < 1 {
		return Book{}, errMissingField
	}
	_, ok := Books[isbn]
	if ok {
		// delete the old entry
		delete(Books, isbn)
		Books[b.ISBN] = b
		return b, nil
	}
	return Book{}, errDoesNotExist
}

// DeleteBook will delete the book from the books list
func DeleteBook(isbn string) (Book, error) {
	b, ok := Books[isbn]
	if ok {
		// delete the entry
		delete(Books, isbn)
		return b, nil
	}
	return Book{}, errDoesNotExist
}
