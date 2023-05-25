package main

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"net/http"
	"os"
	"strings"

	"gopkg.in/fsnotify.v1"

	"github.com/gin-gonic/gin"
)

func Handler(db *pgx.Conn) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := make([]struct {
			IdTitle int    `json:"id_title"`
			Title   string `json:"title"`
			UUID    string `json:"uuid"`
		}, 0)
		const query = "SELECT * FROM title_history WHERE id_title=2"
		err := pgxscan.Select(c, db, &data, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{"hasil": data})
	}
}

func main() {
	r := gin.Default()

	conn, err := pgx.Connect(context.Background(), "DB_URL")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello Rizky",
		})
	})
	r.GET("/db", Handler(conn))

	// creates a new file watcher for App_offline.htm
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	// watch for App_offline.htm and exit the program if present
	// This allows continuous deployment on App Service as the .exe will not be
	// terminated otherwise
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.HasSuffix(event.Name, "App_offline.htm") {
					fmt.Println("Exiting due to App_offline.htm being present")
					os.Exit(0)
				}
			}
		}
	}()

	// get the current working directory and watch it
	currentDir, err := os.Getwd()
	if err := watcher.Add(currentDir); err != nil {
		fmt.Println("ERROR", err)
	}

	port := os.Getenv("HTTP_PLATFORM_PORT")

	if port == "" {
		port = "8080"
	}

	r.Run("127.0.0.1:" + port) // listen and serve on 0.0.0.0:8080
}
