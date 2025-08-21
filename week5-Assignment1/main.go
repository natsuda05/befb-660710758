package main

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

type Product struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    Price    string  `json:"price"`
    Stock    int `json:"stock"`
}

var products = []Product{
    {ID: "1", Name: "Latte", Price: "60.00", Stock: 10},
	{ID: "2", Name: "Espresso", Price: "50.00", Stock: 8},
	{ID: "3", Name: "Americano", Price: "50.00", Stock: 12},
	{ID: "4", Name: "Cappuccino", Price: "60.00", Stock: 6},
}

func getProducts(c *gin.Context){
	priceQuery := c.Query("year")

	if priceQuery != "" {
		filter := []Product{}
        for _, product := range products {
            if fmt.Sprint(product.Price) == priceQuery {
                filter = append(filter, product)
            }
        }
        c.JSON(http.StatusOK, filter)
        return
	}
	c.JSON(http.StatusOK,products)
}

func main()  {
	r := gin.Default()

	r.GET("/health" , func (c *gin.Context)  {
		c.JSON(200, gin.H{"message" : "healthy"})
	})

	api := r.Group("/api/v1")
	{
		api.GET("/products", getProducts)
	}
		
	r.Run(":8080")
}