package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogPost struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id"`
	FileName    string             `bson:"filename"`
	Title       string             `bson:"title"`
	DateTime    string             `bson:"datetime"`
	Tags        []string           `bson:"tags"`
	Views       int32              `bson:"views"`
	ViewIPs     []viewIP           `json:"viewIPs,omitempty" bson:"viewIPs"`
	Content     string             `bson:"content"`
	Description string             `bson:"description"`
}

type viewIP struct {
	IP        string    `bson:"ip"`
	Timestamp time.Time `bson:"timestamp"`
}

func fnConnectMongoDB() (*mongo.Client, error) {
	godotenv.Load()
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
	cur, err := coll.Find(context.TODO(), bson.M{}, options.Find().SetSort(bson.D{{"datetime", -1}}))
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
	var oPost BlogPost
	filter := bson.M{"filename": fileName}
	err := coll.FindOne(context.TODO(), filter).Decode(&oPost)
	if err != nil {
		return nil, err
	}
	return &oPost, nil
}

func fnHasViewedRecently(viewIPs []viewIP, ip string) bool {
	for _, v := range viewIPs {
		if v.IP == ip && time.Since(v.Timestamp) < 24*time.Hour {
			return true
		}
	}
	return false
}

func fnUpdateView(client *mongo.Client, fileName string, ip string) {
	coll := client.Database("general").Collection("blogPosts")
	var oPost BlogPost
	err := coll.FindOne(context.TODO(), bson.M{"filename": fileName}).Decode(&oPost)
	if err != nil {
		log.Fatal(err)
	}
	if !fnHasViewedRecently(oPost.ViewIPs, ip) {
		filter := bson.M{"filename": fileName}
		update := bson.M{
			"$inc":  bson.M{"views": 1},
			"$push": bson.M{"viewIPs": bson.M{"ip": ip, "timestamp": time.Now()}},
		}
		_, err := coll.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Fatal(err)
		}
	}
}
