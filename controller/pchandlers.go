package controller

import (
	"text/template"
	"github.com/gin-gonic/gin"
)

func PCHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc.html"))
	tpl.Execute(c.Writer, nil)
}
