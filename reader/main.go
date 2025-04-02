package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

// simple book struct thingy for testing the database
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

	err = dbPool.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database connection pool successful")

	// RabbitMQ Setup
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

	// gonic/gin setup
	router := gin.Default()
	router.GET("/books/:id", getBook)
	router.GET("/books", listBooks)
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

// simple get by id func
func getBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id")) // id are always int
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// attempt to retrieve the book with the inputted id
	var book Book
	err = dbPool.QueryRow(context.Background(), "SELECT id, title, author FROM test_schema.books WHERE id = $1", id).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// return json with book data if successful
	logEvent("book.get_book_entry", book, c.ClientIP())
	c.JSON(http.StatusOK, book)
}

// returns a list of all books and its data. not suitable for actual system may require pagination.
func listBooks(c *gin.Context) {
	// fetch all rows from the pg
	rows, err := dbPool.Query(context.Background(), "SELECT id, title, author FROM test_schema.books")
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
	logEvent("book.get_all_books", Book{ID: 0}, c.ClientIP())
	c.JSON(http.StatusOK, books)
}
