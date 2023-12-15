
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Deploy endpoint
	r.POST("/deploy", func(c *gin.Context) {
		word := c.PostForm("word")
		// Use the word in your application logic
		fmt.Println("Deploying:", word)

		c.JSON(http.StatusOK, gin.H{
			"message": "Deployment successful",
		})
	})

	r.Run() // Start the server
}
