package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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

var adminSessions = make(map[string]bool)
var adminUsername string
var adminPassword string

func initDB() {
	godotenv.Load()
	
	adminUsername = os.Getenv("ADMIN_USERNAME")
	adminPassword = os.Getenv("ADMIN_PASSWORD")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin123"
	}

	var err error
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL not set in environment")
	}
	
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	log.Println("Database connected successfully")
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
	r.POST("/admin/login", adminLogin)
	r.POST("/admin/logout", adminLogout)
	r.POST("/admin/tv", requireAuth, addTV)
	r.PUT("/admin/tv/:id/start", requireAuth, startTV)
	r.PUT("/admin/tv/:id/stop", requireAuth, stopTV)
	r.GET("/admin/check", checkAuth)

	// Public routes
	r.GET("/", publicPage)
	r.GET("/api/tvs", getTVs)
	r.GET("/api/time", getServerTime)
	r.GET("/ws", websocketHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	r.Run(":" + port)
}

func adminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", nil)
}

func adminLogin(c *gin.Context) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	
	// Check database first
	var dbPassword string
	err := db.QueryRow("SELECT password FROM admins WHERE username = $1", creds.Username).Scan(&dbPassword)
	if err == nil && creds.Password == dbPassword {
		sessionID := strconv.FormatInt(time.Now().UnixNano(), 36)
		adminSessions[sessionID] = true
		c.SetCookie("admin_session", sessionID, 3600*8, "/", "", false, true)
		c.JSON(200, gin.H{"success": true})
		return
	}
	
	// Fallback to env variables
	if creds.Username == adminUsername && creds.Password == adminPassword {
		sessionID := strconv.FormatInt(time.Now().UnixNano(), 36)
		adminSessions[sessionID] = true
		c.SetCookie("admin_session", sessionID, 3600*8, "/", "", false, true)
		c.JSON(200, gin.H{"success": true})
	} else {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
	}
}

func adminLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("admin_session")
	delete(adminSessions, sessionID)
	c.SetCookie("admin_session", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"success": true})
}

func checkAuth(c *gin.Context) {
	sessionID, err := c.Cookie("admin_session")
	if err != nil || !adminSessions[sessionID] {
		c.JSON(401, gin.H{"authenticated": false})
		return
	}
	c.JSON(200, gin.H{"authenticated": true})
}

func requireAuth(c *gin.Context) {
	sessionID, err := c.Cookie("admin_session")
	if err != nil || !adminSessions[sessionID] {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

func getServerTime(c *gin.Context) {
	c.JSON(200, gin.H{"time": time.Now().Unix()})
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
	hours, _ := strconv.Atoi(c.PostForm("duration"))
	duration := hours * 60 // Convert jam ke menit
	
	// Use time from frontend if provided, otherwise server time
	startTimeStr := c.PostForm("start_time")
	endTimeStr := c.PostForm("end_time")
	
	var startTime, endTime time.Time
	var err error
	
	if startTimeStr != "" && endTimeStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startTimeStr)
		endTime, _ = time.Parse(time.RFC3339, endTimeStr)
	} else {
		startTime = time.Now()
		endTime = startTime.Add(time.Duration(duration) * time.Minute)
	}

	_, err = db.Exec(`
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
	rows, err := db.Query("SELECT id, name, status, start_time, duration, end_time FROM tvs ORDER BY id")
	if err != nil {
		log.Printf("Database query error: %v", err)
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var tvs []TV
	for rows.Next() {
		var tv TV
		err := rows.Scan(&tv.ID, &tv.Name, &tv.Status, &tv.StartTime, &tv.Duration, &tv.EndTime)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		
		// Auto update status jika waktu habis
		if tv.Status == "playing" && tv.EndTime != nil && time.Now().After(*tv.EndTime) {
			_, err = db.Exec("UPDATE tvs SET status = 'ready', start_time = NULL, duration = 0, end_time = NULL WHERE id = $1", tv.ID)
			if err == nil {
				tv.Status = "ready"
				tv.StartTime = nil
				tv.EndTime = nil
				tv.Duration = 0
			}
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