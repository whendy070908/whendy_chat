package internal

import (
	"database/sql"
	"strings"
	"time"
)

// ===== WS Protocol =====

type WSIn struct {
	Type      string `json:"type"` // chat|typing|edit|delete
	Content   string `json:"content,omitempty"`
	MessageID int64  `json:"messageId,omitempty"`
	IsTyping  bool   `json:"isTyping,omitempty"`
}

type WSOut struct {
	Type      string        `json:"type"`
	User      string        `json:"user,omitempty"`
	UserID    int64         `json:"userId,omitempty"`
	ChannelID int64         `json:"channelId,omitempty"`
	MessageID int64         `json:"messageId,omitempty"`
	Content   string        `json:"content,omitempty"`
	Timestamp int64         `json:"timestamp,omitempty"`
	History   []MessageRow  `json:"history,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type MessageRow struct {
	ID        int64  `json:"id"`
	ChannelID int64  `json:"channelId"`
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"createdAt"`
}

// ===== DB functions =====

func InsertMessage(db *sql.DB, channelID, userID int64, content string) (int64, int64, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return 0, 0, sql.ErrNoRows
	}

	now := time.Now().UnixMilli()
	res, err := db.Exec(
		`INSERT INTO messages(channel_id,user_id,content,created_at) VALUES(?,?,?,?)`,
		channelID, userID, content, now,
	)
	if err != nil {
		return 0, 0, err
	}
	id, _ := res.LastInsertId()
	return id, now, nil
}

func GetRecentMessages(db *sql.DB, channelID int64, limit int) ([]MessageRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	rows, err := db.Query(`
		SELECT m.id, m.channel_id, m.user_id, u.username, m.content, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.channel_id = ?
		ORDER BY m.id DESC
		LIMIT ?`, channelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tmp := make([]MessageRow, 0, limit)
	for rows.Next() {
		var r MessageRow
		if err := rows.Scan(&r.ID, &r.ChannelID, &r.UserID, &r.Username, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		tmp = append(tmp, r)
	}

	for i, j := 0, len(tmp)-1; i < j; i, j = i+1, j-1 {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp, nil
}

func IsMessageOwner(db *sql.DB, messageID, userID int64) (bool, error) {
	var owner int64
	err := db.QueryRow(`SELECT user_id FROM messages WHERE id = ?`, messageID).Scan(&owner)
	if err != nil {
		return false, err
	}
	return owner == userID, nil
}

func UpdateMessage(db *sql.DB, messageID int64, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return sql.ErrNoRows
	}
	_, err := db.Exec(`UPDATE messages SET content = ? WHERE id = ?`, content, messageID)
	return err
}

func DeleteMessage(db *sql.DB, messageID int64) error {
	_, err := db.Exec(`DELETE FROM messages WHERE id = ?`, messageID)
	return err
}
