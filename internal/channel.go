package internal

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterChannel(r *gin.Engine, db *sql.DB) {
	r.GET("/api/servers/:sid/channels", AuthMiddleware(db), func(c *gin.Context) {
		sid := parseID(c.Param("sid"))
		rows, _ := db.Query(
			"SELECT id,name FROM channels WHERE server_id=?", sid)

		var out []gin.H
		for rows.Next() {
			var id int64
			var name string
			rows.Scan(&id, &name)
			out = append(out, gin.H{"id": id, "name": name})
		}
		c.JSON(200, out)
	})
}
