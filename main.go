package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"whendy_chat/internal"
)

func main() {
	db := internal.OpenDB()
	internal.Migrate(db)

	r := gin.Default()

	internal.RegisterAuth(r, db)
	internal.RegisterServer(r, db)
	internal.RegisterChannel(r, db)
	internal.RegisterWebSocket(r, db)

	log.Println("running on :8080")
	r.Run(":8080")
}
