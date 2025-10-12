package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "week10-lab3/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ---------- Struct ----------

type ErrorResponse struct {
	Message string `json:"message"`
}

type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	ISBN      string    `json:"isbn"`
	Year      int       `json:"year"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ---------- Database ----------

var db *sql.DB

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB() {
	var err error

	host := getEnv("DB_HOST", "localhost")
	name := getEnv("DB_NAME", "bookstore")
	user := getEnv("DB_USER", "bookstore_user")
	password := getEnv("DB_PASSWORD", "your_strong_password")
	port := getEnv("DB_PORT", "5432")

	conSt := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	db, err = sql.Open("postgres", conSt)
	if err != nil {
		log.Fatal("failed to open database")
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Fatal("failed to connect to database")
	}

	log.Println("successfully connect to database")
}

// ---------- Handlers ----------

// @Summary Get all books
// @Description Retrieve all books from the database
// @Tags Books
// @Produce  json
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books [get]
func getAllBooks(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, author, isbn, year, price, created_at, updated_at FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt); err != nil {
			continue
		}
		books = append(books, book)
	}
	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get book by ID
// @Description Retrieve a single book by its ID
// @Tags Books
// @Produce  json
// @Param   id   path   int   true   "Book ID"
// @Success 200  {object}  Book
// @Failure 404  {object}  ErrorResponse
// @Failure 500  {object}  ErrorResponse
// @Router  /books/{id} [get]
func getBook(c *gin.Context) {
	id := c.Param("id")
	var book Book

	err := db.QueryRow("SELECT id, title, author, isbn, year, price, created_at, updated_at FROM books WHERE id = $1", id).
		Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price, &book.CreatedAt, &book.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

// @Summary Create a new book
// @Description Add a new book to the database
// @Tags Books
// @Accept  json
// @Produce  json
// @Param   book  body  Book  true  "Book Data"
// @Success 201  {object}  Book
// @Failure 400  {object}  ErrorResponse
// @Failure 500  {object}  ErrorResponse
// @Router  /books [post]
func createBook(c *gin.Context) {
	var newBook Book
	if err := c.ShouldBindJSON(&newBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id int
	var createdAt, updatedAt time.Time

	err := db.QueryRow(
		`INSERT INTO books (title, author, isbn, year, price)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id, created_at, updated_at`,
		newBook.Title, newBook.Author, newBook.ISBN, newBook.Year, newBook.Price,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newBook.ID = id
	newBook.CreatedAt = createdAt
	newBook.UpdatedAt = updatedAt

	c.JSON(http.StatusCreated, newBook)
}

// @Summary Update an existing book
// @Description Update book details by ID
// @Tags Books
// @Accept  json
// @Produce  json
// @Param   id    path   int   true   "Book ID"
// @Param   book  body   Book  true   "Updated Book Data"
// @Success 200  {object}  Book
// @Failure 400  {object}  ErrorResponse
// @Failure 404  {object}  ErrorResponse
// @Failure 500  {object}  ErrorResponse
// @Router  /books/{id} [put]
func updateBook(c *gin.Context) {
	id := c.Param("id")
	var updateBook Book

	if err := c.ShouldBindJSON(&updateBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedAt time.Time
	err := db.QueryRow(
		`UPDATE books
         SET title = $1, author = $2, isbn = $3, year = $4, price = $5
         WHERE id = $6
         RETURNING updated_at`,
		updateBook.Title, updateBook.Author, updateBook.ISBN, updateBook.Year, updateBook.Price, id,
	).Scan(&updatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updateBook.ID, _ = strconv.Atoi(id)
	updateBook.UpdatedAt = updatedAt

	c.JSON(http.StatusOK, updateBook)
}

// @Summary Delete a book
// @Description Delete a book by its ID
// @Tags Books
// @Produce  json
// @Param   id   path   int   true   "Book ID"
// @Success 200  {object}  map[string]string
// @Failure 404  {object}  ErrorResponse
// @Failure 500  {object}  ErrorResponse
// @Router  /books/{id} [delete]
func deleteBook(c *gin.Context) {
	id := c.Param("id")

	result, err := db.Exec("DELETE FROM books WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book deleted successfully"})
}
func seedDatabase() {

	_, err := db.Exec(`TRUNCATE TABLE books RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Fatalf("Failed to clear table: %v", err)
	}

	log.Println("Table 'books' cleared.")


	booksToSeed := []Book{
		{Title: "The Go Programming Language", Author: "Alan A. A. Donovan", ISBN: "978-0134190440", Year: 2015, Price: 890.50},
		{Title: "Clean Architecture", Author: "Robert C. Martin", ISBN: "978-0134494166", Year: 2017, Price: 1250.00},
		{Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", ISBN: "978-1449373320", Year: 2017, Price: 1500.75},
	}

	
	for _, book := range booksToSeed {
		_, err := db.Exec(`
			INSERT INTO books (title, author, isbn, year, price)
			VALUES ($1, $2, $3, $4, $5)`,
			book.Title, book.Author, book.ISBN, book.Year, book.Price)

		if err != nil {
			log.Fatalf("Failed to seed book '%s': %v", book.Title, err)
		}
	}

	log.Println("Database seeded with initial data successfully!")
}
// ---------- Swagger Info ----------

// @title           Bookstore API Example
// @version         1.0
// @description     This is a simple example API for managing books.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	initDB()
	seedDatabase()
	defer db.Close()

	r := gin.Default()
	r.Use(cors.Default())

	// Swagger docs route
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unhealthy", "err": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.GET("/books", getAllBooks)
		api.GET("/books/:id", getBook)
		api.POST("/books", createBook)
		api.PUT("/books/:id", updateBook)
		api.DELETE("/books/:id", deleteBook)
	}

	r.Run(":8080")
}
