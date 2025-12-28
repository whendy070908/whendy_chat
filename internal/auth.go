package internal

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const JWTKey = "CHANGE_THIS_SECRET"

type User struct {
	ID       int64
	Username string
}

func RegisterAuth(r *gin.Engine, db *sql.DB) {
	r.POST("/api/auth/signup", func(c *gin.Context) {
		var b struct{ Username, Password string }
		c.BindJSON(&b)

		res, err := db.Exec(
			"INSERT INTO users(username,password) VALUES(?,?)",
			b.Username, b.Password,
		)
		if err != nil {
			c.JSON(409, gin.H{"error": "exists"})
			return
		}

		uid, _ := res.LastInsertId()
		c.JSON(200, gin.H{"token": signJWT(uid, b.Username)})
	})

	r.POST("/api/auth/login", func(c *gin.Context) {
		var b struct{ Username, Password string }
		c.BindJSON(&b)

		var uid int64
		err := db.QueryRow(
			"SELECT id FROM users WHERE username=? AND password=?",
			b.Username, b.Password,
		).Scan(&uid)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid"})
			return
		}

		c.JSON(200, gin.H{"token": signJWT(uid, b.Username)})
	})
}

func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatus(401)
			return
		}
		u, err := verifyJWT(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		c.Set("user", u)
		c.Next()
	}
}

func signJWT(uid int64, username string) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		"u":   username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})
	s, _ := t.SignedString([]byte(JWTKey))
	return s
}

func verifyJWT(tok string) (User, error) {
	t, err := jwt.Parse(tok, func(t *jwt.Token) (any, error) {
		return []byte(JWTKey), nil
	})
	if err != nil || !t.Valid {
		return User{}, err
	}
	c := t.Claims.(jwt.MapClaims)
	return User{
		ID:       int64(c["uid"].(float64)),
		Username: c["u"].(string),
	}, nil
}
