package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	// ต้องสร้างไฟล์ docs/docs.go และรัน swaggo init ก่อนจึงจะใช้ได้
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ---------- Struct ----------

type ErrorResponse struct {
	Message string `json:"message"`
}

type Book struct {
	ID              int        `json:"id"`
	Title           string     `json:"title" binding:"required"`
	Author          string     `json:"author" binding:"required"`
	ISBN            string     `json:"isbn" binding:"required"`
	Year            int        `json:"year" binding:"required"`
	Price           float64    `json:"price" binding:"required"`

	// ฟิลด์ใหม่ 13 ฟิลด์
	Category        string     `json:"category" binding:"required"`
	OriginalPrice   *float64   `json:"original_price"` // ใช้ Pointer สำหรับ Nullable fields
	Discount        int        `json:"discount"`
	CoverImage      string     `json:"cover_image" binding:"required"`
	Rating          float64    `json:"rating"`
	ReviewsCount    int        `json:"reviews_count"`
	IsNew           bool       `json:"is_new"`
	Pages           *int       `json:"pages"` // ใช้ Pointer สำหรับ Nullable fields
	Language        string     `json:"language" binding:"required"`
	Publisher       string     `json:"publisher" binding:"required"`
	Description     string     `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Global variable สำหรับ Column List
const bookColumns = "id, title, author, isbn, year, price, category, original_price, discount, cover_image, rating, reviews_count, is_new, pages, language, publisher, description, created_at, updated_at"

// ฟังก์ชัน Helper ในการ Scan Row ใหม่
func scanBook(row *sql.Row, book *Book) error {
	return row.Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price,
		&book.Category, &book.OriginalPrice, &book.Discount, &book.CoverImage,
		&book.Rating, &book.ReviewsCount, &book.IsNew, &book.Pages, &book.Language,
		&book.Publisher, &book.Description, &book.CreatedAt, &book.UpdatedAt,
	)
}

func scanBookRows(rows *sql.Rows, book *Book) error {
	return rows.Scan(
		&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Year, &book.Price,
		&book.Category, &book.OriginalPrice, &book.Discount, &book.CoverImage,
		&book.Rating, &book.ReviewsCount, &book.IsNew, &book.Pages, &book.Language,
		&book.Publisher, &book.Description, &book.CreatedAt, &book.UpdatedAt,
	)
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
		log.Fatal("failed to open database:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("successfully connect to database")
}

// ---------- Handlers ----------

// @Summary Get all books
// @Description Retrieve all books from the database. Can filter by category.
// @Tags Books
// @Produce  json
// @Param   category query string false "Filter by category"
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books [get]
func getAllBooks(c *gin.Context) {
	category := c.Query("category")
	
	selectQuery := fmt.Sprintf("SELECT %s FROM books", bookColumns)
	args := []interface{}{}
	
	if category != "" {
		selectQuery += " WHERE category = $1"
		args = append(args, category)
	}

	rows, err := db.Query(selectQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := scanBookRows(rows, &book); err != nil {
			log.Printf("Error scanning row: %v", err)
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

	row := db.QueryRow(fmt.Sprintf("SELECT %s FROM books WHERE id = $1", bookColumns), id)

	if err := scanBook(row, &book); err == sql.ErrNoRows {
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

	insertColumns := "title, author, isbn, year, price, category, original_price, discount, cover_image, rating, reviews_count, is_new, pages, language, publisher, description"
	valuePlaceholders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17"

	var id int
	var createdAt, updatedAt time.Time

	err := db.QueryRow(
		fmt.Sprintf(`INSERT INTO books (%s) VALUES (%s) RETURNING id, created_at, updated_at`, insertColumns, valuePlaceholders),
		newBook.Title, newBook.Author, newBook.ISBN, newBook.Year, newBook.Price,
		newBook.Category, newBook.OriginalPrice, newBook.Discount, newBook.CoverImage,
		newBook.Rating, newBook.ReviewsCount, newBook.IsNew, newBook.Pages, newBook.Language,
		newBook.Publisher, newBook.Description,
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

	updateSet := `
		title = $1, author = $2, isbn = $3, year = $4, price = $5,
		category = $6, original_price = $7, discount = $8, cover_image = $9,
		rating = $10, reviews_count = $11, is_new = $12, pages = $13,
		language = $14, publisher = $15, description = $16
	`
	
	var updatedAt time.Time
	err := db.QueryRow(
		fmt.Sprintf(`UPDATE books SET %s WHERE id = $17 RETURNING updated_at`, updateSet),
		updateBook.Title, updateBook.Author, updateBook.ISBN, updateBook.Year, updateBook.Price,
		updateBook.Category, updateBook.OriginalPrice, updateBook.Discount, updateBook.CoverImage,
		updateBook.Rating, updateBook.ReviewsCount, updateBook.IsNew, updateBook.Pages,
		updateBook.Language, updateBook.Publisher, updateBook.Description,
		id,
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

// @Summary Get unique categories
// @Description Retrieve a list of all unique book categories
// @Tags Categories
// @Produce  json
// @Success 200  {array}  string
// @Failure 500  {object}  ErrorResponse
// @Router  /categories [get]
func getCategories(c *gin.Context) {
	rows, err := db.Query("SELECT DISTINCT category FROM books ORDER BY category")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			continue
		}
		categories = append(categories, category)
	}

	c.JSON(http.StatusOK, categories)
}

// @Summary Search books
// @Description Search books by title, author, or description
// @Tags Books
// @Produce  json
// @Param   q query string true "Search keyword for title, author, or description"
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books/search [get]
func searchBooks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' is required"})
		return
	}

	searchPattern := "%" + strings.ToLower(query) + "%"
	
	selectQuery := fmt.Sprintf("SELECT %s FROM books WHERE LOWER(title) LIKE $1 OR LOWER(author) LIKE $1 OR LOWER(description) LIKE $1", bookColumns)
	
	rows, err := db.Query(selectQuery, searchPattern)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := scanBookRows(rows, &book); err != nil {
			log.Printf("Error scanning search result: %v", err)
			continue
		}
		books = append(books, book)
	}

	if books == nil {
		books = []Book{}
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get featured books
// @Description Retrieve books with high ratings
// @Tags Books
// @Produce  json
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books/featured [get]
func getFeaturedBooks(c *gin.Context) {
	// ตัวอย่าง: ดึงหนังสือที่มี Rating มากกว่าหรือเท่ากับ 4.5
	selectQuery := fmt.Sprintf("SELECT %s FROM books WHERE rating >= 4.5 ORDER BY rating DESC, reviews_count DESC LIMIT 10", bookColumns)
	
	rows, err := db.Query(selectQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := scanBookRows(rows, &book); err != nil {
			log.Printf("Error scanning featured book: %v", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get new books
// @Description Retrieve books marked as new or recently added
// @Tags Books
// @Produce  json
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books/new [get]
func getNewBooks(c *gin.Context) {
	// ตัวอย่าง: ดึงหนังสือที่ IsNew เป็น true หรือ CreatedAt ภายใน 30 วันล่าสุด
	selectQuery := fmt.Sprintf("SELECT %s FROM books WHERE is_new = TRUE OR created_at >= (NOW() - INTERVAL '30 DAYS') ORDER BY created_at DESC LIMIT 10", bookColumns)
	
	rows, err := db.Query(selectQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := scanBookRows(rows, &book); err != nil {
			log.Printf("Error scanning new book: %v", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// @Summary Get discounted books
// @Description Retrieve books that are currently on discount
// @Tags Books
// @Produce  json
// @Success 200  {array}  Book
// @Failure 500  {object}  ErrorResponse
// @Router  /books/discounted [get]
func getDiscountedBooks(c *gin.Context) {
	// ตัวอย่าง: ดึงหนังสือที่มี Discount > 0
	selectQuery := fmt.Sprintf("SELECT %s FROM books WHERE discount > 0 ORDER BY discount DESC, title ASC LIMIT 20", bookColumns)
	
	rows, err := db.Query(selectQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := scanBookRows(rows, &book); err != nil {
			log.Printf("Error scanning discounted book: %v", err)
			continue
		}
		books = append(books, book)
	}
	c.JSON(http.StatusOK, books)
}

// ---------- Seeder ----------
func seedDatabase() {

	_, err := db.Exec(`TRUNCATE TABLE books RESTART IDENTITY CASCADE`)
	if err != nil {
		log.Fatalf("Failed to clear table: %v", err)
	}
	log.Println("Table 'books' cleared.")

	// ตัวอย่างข้อมูลสำหรับฟิลด์ใหม่ (ใช้ *float64 และ *int สำหรับฟิลด์ nullable)
	originalPrice1 := 990.00
	pages1 := 500
	
	originalPrice2 := 1600.00
	pages2 := 800
	
	originalPrice3 := 1500.75
	pages3 := 650

	booksToSeed := []Book{
		{
			Title: "The Go Programming Language", Author: "Alan A. A. Donovan", ISBN: "978-0134190440", Year: 2015, Price: 890.50,
			Category: "Programming", OriginalPrice: &originalPrice1, Discount: 10, CoverImage: "go.jpg",
			Rating: 4.8, ReviewsCount: 150, IsNew: false, Pages: &pages1, Language: "English",
			Publisher: "Addison-Wesley Professional", Description: "A comprehensive guide to the Go language.",
		},
		{
			Title: "Clean Architecture", Author: "Robert C. Martin", ISBN: "978-0134494166", Year: 2017, Price: 1250.00,
			Category: "Software Design", OriginalPrice: &originalPrice2, Discount: 21, CoverImage: "clean.jpg",
			Rating: 4.5, ReviewsCount: 90, IsNew: true, Pages: &pages2, Language: "English",
			Publisher: "Prentice Hall", Description: "A blueprint for software structure.",
		},
		{
			Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", ISBN: "978-1449373320", Year: 2017, Price: 1500.75,
			Category: "Database", OriginalPrice: &originalPrice3, Discount: 0, CoverImage: "data.jpg",
			Rating: 4.9, ReviewsCount: 200, IsNew: false, Pages: &pages3, Language: "English",
			Publisher: "O'Reilly Media", Description: "The essential guide to the fundamentals of systems.",
		},
	}

	insertColumns := "title, author, isbn, year, price, category, original_price, discount, cover_image, rating, reviews_count, is_new, pages, language, publisher, description"
	valuePlaceholders := "$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16"
	
	for _, book := range booksToSeed {
		_, err := db.Exec(
			fmt.Sprintf(`INSERT INTO books (%s) VALUES (%s)`, insertColumns, valuePlaceholders),
			book.Title, book.Author, book.ISBN, book.Year, book.Price,
			book.Category, book.OriginalPrice, book.Discount, book.CoverImage,
			book.Rating, book.ReviewsCount, book.IsNew, book.Pages, book.Language,
			book.Publisher, book.Description,
		)

		if err != nil {
			log.Fatalf("Failed to seed book '%s': %v", book.Title, err)
		}
	}

	log.Println("Database seeded with initial data successfully!")
}

// ---------- Swagger Info (ต้องใช้ go get -u github.com/swaggo/swag/cmd/swag และรัน swag init) ----------

// @title           Bookstore API Example (Extended)
// @version         1.0
// @description     This is an extended API for managing books with rich data.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	initDB()
	seedDatabase() // รัน Seed Data ทุกครั้งที่ Server เริ่มต้น
	defer db.Close()

	r := gin.Default()
	r.Use(cors.Default())

	// Swagger docs route - รัน 'swag init' ก่อนรันโปรแกรม
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": "unhealthy", "err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Basic CRUD Operations
		api.GET("/books", getAllBooks)        // รองรับ /books?category=...
		api.GET("/books/:id", getBook)
		api.POST("/books", createBook)
		api.PUT("/books/:id", updateBook)
		api.DELETE("/books/:id", deleteBook)

		// New Endpoints (Discovery/Search)
		api.GET("/categories", getCategories)
		api.GET("/books/search", searchBooks)
		api.GET("/books/featured", getFeaturedBooks)
		api.GET("/books/new", getNewBooks)
		api.GET("/books/discounted", getDiscountedBooks)
	}

	// เริ่มต้น Server
	log.Println("Server is running on http://localhost:8080")
	r.Run(":8080")
}