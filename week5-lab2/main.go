package main

import (
	"github.com/gin-gonic/gin"
)

func getUsers(c *gin.Context) {
	user := []User{{ID: "1", Name: "Nuttachot"}}

	c.JSON(200, user)
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	r := gin.Default()

	r.GET("/users", getUsers)

	r.Run(":8080")
}
