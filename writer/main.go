package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var db *pgx.Conn
var rabbitConn *amqp.Connection
var rabbitChan *amqp.Channel

func main() {
	var err error

	//PostgreSQL setup
	connStr := "postgres://admin:12345@write-db:5433/testdb?sslmode=disable" //NOSONAR
	db, err = pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection successful")

	// RabbitMQ setup
	rabbitConn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err = rabbitConn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer rabbitChan.Close()

	err = rabbitChan.ExchangeDeclare(
		"book_events", // Exchange name
		"topic",       // Exchange type (topic)
		true,          // Durable
		false,         // Auto-deleted
		false,         // Internal
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	// Gin Setup
	router := gin.Default()

	router.POST("/books", createBook)
	router.PUT("/books/:id", updateBook)
	router.DELETE("/books/:id", deleteBook)

	router.Run(":8080")
}

func logEvent(routingKey string, book Book, sourceIp string) {
	event := map[string]interface{}{
		"book":   book,
		"source": sourceIp,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Println("Error marshalling event:", err)
		return
	}

	err = rabbitChan.Publish(
		"book_events", // Exchange
		routingKey,    // Routing key
		false,         // Mandatory
		false,         // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        eventJSON,
		})
	if err != nil {
		log.Println("Error publishing message:", err)
	} else {
		log.Println("Event published:", routingKey)
	}
}

func createBook(c *gin.Context) {
	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := db.QueryRow(context.Background(), "INSERT INTO test_schema.books (title, author) VALUES ($1, $2) RETURNING id", book.Title, book.Author).Scan(&book.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logEvent("book.created", book, c.ClientIP())
	c.JSON(http.StatusCreated, book)
}

func updateBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	book.ID = id

	ct, err := db.Exec(context.Background(), "UPDATE test_schema.books SET title = $1, author = $2 WHERE id = $3", book.Title, book.Author, book.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	logEvent("book.updated", book, c.ClientIP())
	c.JSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ct, err := db.Exec(context.Background(), "DELETE FROM test_schema.books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	logEvent("book.deleted", Book{ID: id}, c.ClientIP())
	c.JSON(http.StatusNoContent, nil)
}
