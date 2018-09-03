package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/schollz/messagebox/keypair"
	"github.com/schollz/messagebox/messagebox"
)

var world keypair.KeyPair

func init() {
	world, _ := keypair.NewDeterministic("world1")
	log.Println(world)
}

func main() {
	router := gin.Default()

	router.GET("/add/:hash", func(c *gin.Context) {
		hash := c.Param("hash")

		r, err := http.Get("https://ipfs.io/ipfs/" + hash)
		if err != nil {
			c.String(http.StatusOK, err.Error())
			return
		}
		defer r.Body.Close()

		var msg messagebox.Message
		err = json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			c.String(http.StatusOK, err.Error())
		} else {
			log.Println(msg.IsSameWorld(world))
			c.JSON(200, msg)
		}
	})

	router.GET("/all", func(c *gin.Context) {
		// return list of all hashes
		c.String(http.StatusOK, "ok")
	})

	router.Run(":8080")
}
