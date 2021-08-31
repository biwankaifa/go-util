package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var databaseOfMongo *mongo.Database

type ConfigOfMongo struct {
	Host     string
	Port     int
	Name     string
	Pass     string
	DataBase string
}

func InitMongo(cfg ConfigOfMongo) error {
	if databaseOfMongo != nil {
		return nil
	}
	var param string
	param = fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Name, cfg.Pass, cfg.Host, cfg.Port)
	clientOptions := options.Client().ApplyURI(param)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	databaseOfMongo = client.Database(cfg.DataBase)

	return nil
}

func GetMongoDatabase() *mongo.Database {
	return databaseOfMongo
}
