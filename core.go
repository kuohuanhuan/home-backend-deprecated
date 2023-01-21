package main

import (
	"context"
	"log"
	"strings"
	"time"

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
		for i, oPost := range arrPosts {
			oPost.Content = strings.Replace(oPost.Content, "\r\n", "\n", -1)
			oPost.Content = strings.Trim(oPost.Content, "\n")
			arrPosts[i] = oPost
		}
		return c.JSON(arrPosts)
	})
	app.Get("/post/:file", func(c *fiber.Ctx) error {
		if c.Cookies(c.Params("file")) == "" {
			c.Cookie(&fiber.Cookie{
				Name:    c.Params("file"),
				Value:   "You have viewed this post in 24hrs :3",
				Expires: time.Now().Add(24 * time.Hour),
			})
			fnUpdateView(client, c.Params("file"))
		}
		oPost, err := fnGetPost(client, c.Params("file"))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error":   "Not Found",
				"message": "The requested URL was not found on this server. That's all we know.",
			})
		}
		oPost.Content = strings.Replace(oPost.Content, "\r\n", "\n", -1)
		oPost.Content = strings.Trim(oPost.Content, "\n")
		return c.JSON(oPost)
	})
	app.Listen(":80")
}
