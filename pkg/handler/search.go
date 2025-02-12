package handler

import (
	"fmt"
	"net/http"
	"vectorchat/pkg"

	"github.com/gin-gonic/gin"
)

func SearchHandler(c *gin.Context) {

	var data interface{}
	// var request map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var v map[string]interface{}
	switch temp := data.(type) {
	case map[string]interface{}:
		v = temp // Assign temp to v
		fmt.Println(v)
	}

	fmt.Println("data", data)

	result, err := pkg.SearchOpenAI(v["query"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("a2")

	c.JSON(http.StatusOK, gin.H{"result": result})
}
