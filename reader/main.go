package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// simple book struct thingy for testing the database
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var db *pgx.Conn // init the postgres pgx connection object

func main() {
	var err error

	// init connection to pg
	connStr := "postgres://admin:12345@read-db:5432/testdb?sslmode=disable" //NOSONAR
	db, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	// check that pg is succesfully connected
	err = db.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection successful")

	// gonic/gin setup
	router := gin.Default()
	router.GET("/books/:id", getBook)
	router.GET("/books", listBooks)
	router.Run(":8080")
}

// simple get by id func
func getBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")) // id are always int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// attempt to retrieve the book with the inputted id
	var book Book
	err = db.QueryRow(context.Background(), "SELECT id, title, author FROM books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// return json with book data if successful
	c.JSON(http.StatusOK, book)
}

// returns a list of all books and its data. not suitable for actual system may require pagination.
func listBooks(c *gin.Context) {
	// fetch all rows from the pg
	rows, err := db.Query(context.Background(), "SELECT id, title, author FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// populate a slice with the rows data
	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books = append(books, book)
	}

	// returns a json response with the data
	c.JSON(http.StatusOK, books)
}
