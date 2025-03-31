package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.GET("/", rootHandler)
	r.GET("/movies", getMovies)
	r.GET("/movies/:id", getMovie)
	r.GET("/mshealth", healthCheck)
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "There is nothing to do here! 404!"})
	})
	r.Run(":4567")
}

func rootHandler(c *gin.Context) {
	colorGen := NewColorGenerator()
	color := colorGen.CreateHex()

	aphorisms, err := readAphorisms("aphorisms.txt")

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		aphorisms = []string{
			"If I stop to kick every barking dog I am not going to get where I'm going. - Jackie Joyner-Kersee",
			"Optimism is the faith that leads to achievement. - Helen Keller",
		}
	}
	randomAphorism := getRandomAphorism(aphorisms)
	message, author := formatAphorism(randomAphorism)
	html := fmt.Sprintf(`
        <p style="text-align: center; font-size: 2em; color: %s">
            %s <br> - <span style="font-style: italic;">%s</span>
        </p>
    `, color, message, author)
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func getMovies(c *gin.Context) {
	movies, err := readMoviesData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read movies data"})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func getMovie(c *gin.Context) {
	id := c.Param("id")
	movies, err := readMoviesData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read movies data"})
		return
	}

	for _, movie := range movies {
		if movie.ImdbID == id {
			c.JSON(http.StatusOK, movie)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Movie with requested ID not found"})
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
