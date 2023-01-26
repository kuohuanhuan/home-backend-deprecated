package main

import (
	"context"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func fnServer() {
	client, err := fnConnectMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				return c.Status(404).JSON(fiber.Map{
					"error":   "Not Found",
					"message": "The requested URL was not found on this server. That's all we know.",
				})
			}
			return c.Status(500).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "An error occurred while processing your request. That's not your fault.",
			})
		}
		return nil
	})
	app.Use(cors.New())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World! System is running.")
	})
	app.Get("/posts", func(c *fiber.Ctx) error {
		arrPosts, err := fnGetAllPosts(client)
		if err != nil {
			log.Fatal(err)
		}
		arrNewPosts := make([]BlogPost, len(arrPosts))
		for i, oPost := range arrPosts {
			oPost.Content = strings.Replace(oPost.Content, "\r\n", "\n", -1)
			oPost.Content = strings.Trim(oPost.Content, "\n")
			arrNewPosts[i] = BlogPost{
				ID:          oPost.ID,
				FileName:    oPost.FileName,
				Title:       oPost.Title,
				DateTime:    oPost.DateTime,
				Tags:        oPost.Tags,
				Content:     oPost.Content,
				Description: oPost.Description,
				Views:       oPost.Views,
			}
		}
		return c.JSON(arrNewPosts)
	})
	app.Get("/post/:file", func(c *fiber.Ctx) error {
		oPost, err := fnGetPost(client, c.Params("file"))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":   "Internal Server Error",
				"message": "An error occurred while processing your request. That's not your fault.",
			})
		}
		var sIP string = c.Get("X-Forwarded-For")
		if sIP == "" {
			sIP = c.IP()
		}
		fnUpdateView(client, c.Params("file"), sIP)
		oPost.Content = strings.Replace(oPost.Content, "\r\n", "\n", -1)
		oPost.Content = strings.Trim(oPost.Content, "\n")
		return c.JSON(struct {
			BlogPost
			ViewIPs []string `json:"viewIPs,omitempty" bson:"viewIPs"`
		}{BlogPost: *oPost})
	})
	app.Listen(":80")
}
