package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type TV struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"` // ready, playing
	StartTime *time.Time `json:"start_time"`
	Duration  int       `json:"duration"` // dalam menit
	EndTime   *time.Time `json:"end_time"`
}

var db *sql.DB
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func initDB() {
	var err error
	db, err = sql.Open("postgres", "postgresql://fuelfriendly:fuelfriendly123@72.61.69.116:5432/fuelfriendly")
	if err != nil {
		log.Fatal(err)
	}

	// Create table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tvs (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			status VARCHAR(20) DEFAULT 'ready',
			start_time TIMESTAMP,
			duration INTEGER DEFAULT 0,
			end_time TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()
	
	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Admin routes
	r.GET("/admin", adminPage)
	r.POST("/admin/tv", addTV)
	r.PUT("/admin/tv/:id/start", startTV)
	r.PUT("/admin/tv/:id/stop", stopTV)

	// Public routes
	r.GET("/", publicPage)
	r.GET("/api/tvs", getTVs)
	r.GET("/ws", websocketHandler)

	r.Run(":8080")
}

func adminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

func publicPage(c *gin.Context) {
	c.HTML(http.StatusOK, "public.html", nil)
}

func addTV(c *gin.Context) {
	var tv TV
	if err := c.ShouldBindJSON(&tv); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec("INSERT INTO tvs (name) VALUES ($1)", tv.Name)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "TV added successfully"})
	broadcastTVs()
}

func startTV(c *gin.Context) {
	id := c.Param("id")
	duration, _ := strconv.Atoi(c.PostForm("duration"))
	
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	_, err := db.Exec(`
		UPDATE tvs SET status = 'playing', start_time = $1, duration = $2, end_time = $3 
		WHERE id = $4
	`, startTime, duration, endTime, id)
	
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "TV started"})
	broadcastTVs()
}

func stopTV(c *gin.Context) {
	id := c.Param("id")
	
	_, err := db.Exec(`
		UPDATE tvs SET status = 'ready', start_time = NULL, duration = 0, end_time = NULL 
		WHERE id = $1
	`, id)
	
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "TV stopped"})
	broadcastTVs()
}

func getTVs(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, status, start_time, duration, end_time FROM tvs")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tvs []TV
	for rows.Next() {
		var tv TV
		err := rows.Scan(&tv.ID, &tv.Name, &tv.Status, &tv.StartTime, &tv.Duration, &tv.EndTime)
		if err != nil {
			continue
		}
		
		// Auto update status jika waktu habis
		if tv.Status == "playing" && tv.EndTime != nil && time.Now().After(*tv.EndTime) {
			db.Exec("UPDATE tvs SET status = 'ready', start_time = NULL, duration = 0, end_time = NULL WHERE id = $1", tv.ID)
			tv.Status = "ready"
			tv.StartTime = nil
			tv.EndTime = nil
			tv.Duration = 0
		}
		
		tvs = append(tvs, tv)
	}

	c.JSON(200, tvs)
}

var clients = make(map[*websocket.Conn]bool)

func websocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clients[conn] = true
	defer delete(clients, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func broadcastTVs() {
	rows, _ := db.Query("SELECT id, name, status, start_time, duration, end_time FROM tvs")
	defer rows.Close()

	var tvs []TV
	for rows.Next() {
		var tv TV
		rows.Scan(&tv.ID, &tv.Name, &tv.Status, &tv.StartTime, &tv.Duration, &tv.EndTime)
		tvs = append(tvs, tv)
	}

	data, _ := json.Marshal(tvs)
	for client := range clients {
		client.WriteMessage(websocket.TextMessage, data)
	}
}