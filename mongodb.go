package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogPost struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	FileName string             `bson:"fileName"`
	Title    string             `bson:"title"`
	Date     string             `bson:"date"`
	Tags     string             `bson:"tags"`
	Views    int32              `bson:"views"`
	Content  string             `bson:"content"`
}

func fnConnectMongoDB() (*mongo.Client, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	return client, nil
}

func fnGetAllPosts(c *mongo.Client) ([]BlogPost, error) {
	var arrPosts []BlogPost
	coll := c.Database("general").Collection("blogPosts")
	cur, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var oPost BlogPost
		err := cur.Decode(&oPost)
		if err != nil {
			log.Fatal(err)
		}
		arrPosts = append(arrPosts, oPost)
	}
	return arrPosts, nil
}

func fnGetPost(client *mongo.Client, fileName string) (*BlogPost, error) {
	coll := client.Database("general").Collection("blogPosts")
	var blogPost BlogPost
	filter := bson.M{"fileName": fileName}
	err := coll.FindOne(context.TODO(), filter).Decode(&blogPost)
	if err != nil {
		return nil, err
	}
	return &blogPost, nil
}

func fnUpdateView(client *mongo.Client, fileName string) {
	coll := client.Database("general").Collection("blogPosts")
	rand.Seed(time.Now().UnixNano())
	filter := bson.M{"fileName": fileName}
	_, err := coll.UpdateOne(context.TODO(), filter, bson.M{"$inc": bson.M{"views": 1}})
	if err != nil {
		log.Fatal(err)
	}
}
