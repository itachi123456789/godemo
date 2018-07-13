package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProductHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "product.html", nil)
}
