package main

import (
	"database/sql"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
)

func urls(c *gin.Context) {
	db, _ := sql.Open("mysql", os.Getenv("dbConnectString"))

	request := ShortUrlRequest{}
	c.BindJSON(&request)

	_, err := url.ParseRequestURI(request.Url)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "url invalid",
		})
		return
	}

	var expiredAt time.Time
	var res sql.Result

	if request.ExpireAt != "" {
		expiredAt, err = time.Parse(time.RFC3339, request.ExpireAt)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "expiredAt invalid",
			})
		}
	}
	slug := RandString(3)

	if request.ExpireAt != "" {
		stmt, _ := db.Prepare("INSERT INTO records set url=?,slug=?, expired_at=?;")
		res, _ = stmt.Exec(request.Url, slug, expiredAt)
	} else {
		stmt, _ := db.Prepare("INSERT INTO records set url=?,slug=?;")
		res, _ = stmt.Exec(request.Url, slug)
	}

	lastId, _ := res.LastInsertId()

	c.JSON(200, gin.H{
		"id":       strconv.FormatInt(lastId, 10) + slug,
		"shortUrl": os.Getenv("baseUrl") + strconv.FormatInt(lastId, 10) + slug,
	})

	db.Close()
}

func goUrl(c *gin.Context) {
	var url string

	db, _ := sql.Open("mysql", os.Getenv("dbConnectString"))

	row := db.QueryRow("SELECT url FROM records WHERE CONCAT(id, slug)=? and (expired_at >= now() or expired_at is null);", c.Param("id"))
	row.Scan(&url)

	if url == "" {
		c.String(404, "Not Found")
		return
	}

	c.Redirect(302, url)

	db.Close()
}
