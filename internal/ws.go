package internal

import (
	"database/sql"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn      *websocket.Conn
	send      chan any
	user      User
	channelID int64
}

type Room struct {
	clients map[*Client]bool
	mu      sync.Mutex
}

var rooms = map[int64]*Room{}
var roomsMu sync.Mutex

func getRoom(cid int64) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if rooms[cid] == nil {
		rooms[cid] = &Room{clients: map[*Client]bool{}}
	}
	return rooms[cid]
}

func RegisterWebSocket(r *gin.Engine, db *sql.DB) {
	r.GET("/ws", func(c *gin.Context) {
		token := c.Query("token")
		sid, _ := strconv.ParseInt(c.Query("server_id"), 10, 64)
		cid, _ := strconv.ParseInt(c.Query("channel_id"), 10, 64)

		u, err := verifyJWT(token)
		if err != nil || !IsServerMember(db, u.ID, sid) {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		conn, _ := upgrader.Upgrade(c.Writer, c.Request, nil)
		client := &Client{
			conn:      conn,
			send:      make(chan any),
			user:      u,
			channelID: cid,
		}

		room := getRoom(cid)
		room.mu.Lock()
		room.clients[client] = true
		room.mu.Unlock()

		go writePump(client)
		readPump(db, room, client)
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func readPump(db *sql.DB, room *Room, c *Client) {
	defer func() {
		room.mu.Lock()
		delete(room.clients, c)
		room.mu.Unlock()
		c.conn.Close()
	}()

	for {
		var msg struct{ Content string }
		if err := c.conn.ReadJSON(&msg); err != nil {
			return
		}

		db.Exec(
			"INSERT INTO messages(channel_id,user_id,content,created_at) VALUES(?,?,?,?)",
			c.channelID, c.user.ID, msg.Content, time.Now().UnixMilli(),
		)

		room.mu.Lock()
		for cl := range room.clients {
			cl.send <- gin.H{
				"user":    c.user.Username,
				"content": msg.Content,
			}
		}
		room.mu.Unlock()
	}
}

func writePump(c *Client) {
	for m := range c.send {
		c.conn.WriteJSON(m)
	}
}
