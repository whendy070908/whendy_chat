package internal

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterServer(r *gin.Engine, db *sql.DB) {
	r.POST("/api/servers", AuthMiddleware(db), func(c *gin.Context) {
		u := c.MustGet("user").(User)
		var b struct{ Name string }
		c.BindJSON(&b)

		tx, _ := db.Begin()
		res, _ := tx.Exec(
			"INSERT INTO servers(name,owner_id,created_at) VALUES(?,?,?)",
			b.Name, u.ID, time.Now().UnixMilli(),
		)
		sid, _ := res.LastInsertId()
		tx.Exec("INSERT INTO server_members VALUES(?,?,?)", sid, u.ID, "owner")
		tx.Exec("INSERT INTO channels(server_id,name,created_at) VALUES(?,?,?)",
			sid, "general", time.Now().UnixMilli())
		tx.Commit()

		c.JSON(200, gin.H{"id": sid})
	})

	r.GET("/api/servers", AuthMiddleware(db), func(c *gin.Context) {
		u := c.MustGet("user").(User)
		rows, _ := db.Query(`
			SELECT s.id, s.name FROM servers s
			JOIN server_members m ON m.server_id=s.id
			WHERE m.user_id=?`, u.ID)

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

func IsServerMember(db *sql.DB, uid, sid int64) bool {
	var cnt int
	db.QueryRow(
		"SELECT COUNT(*) FROM server_members WHERE user_id=? AND server_id=?",
		uid, sid,
	).Scan(&cnt)
	return cnt > 0
}

func parseID(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
