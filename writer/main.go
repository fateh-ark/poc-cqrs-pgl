package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var dbPool *pgxpool.Pool // Use pgxpool.Pool
var rabbitConn *amqp.Connection
var rabbitChan *amqp.Channel

func main() {
	var err error

	// PostgreSQL setup using pgxpool
	connStr := "postgres://admin:12345@pgpool:5432/testdb?sslmode=disable" //NOSONAR
	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		dbPool, err = pgxpool.New(context.Background(), connStr)
		if err == nil {
			break // Connection successful
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	defer dbPool.Close() // Close the pool

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

	err := dbPool.QueryRow(context.Background(), "INSERT INTO test_schema.books (title, author) VALUES ($1, $2) RETURNING id", book.Title, book.Author).Scan(&book.ID)
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

	ct, err := dbPool.Exec(context.Background(), "UPDATE test_schema.books SET title = $1, author = $2 WHERE id = $3", book.Title, book.Author, book.ID)

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

	ct, err := dbPool.Exec(context.Background(), "DELETE FROM test_schema.books WHERE id = $1", id)
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
