package mongo

import (
	"context"
	"errors"
	"fmt"
	"go-util/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MaxConnectionNum = 500

var cfg = config.Get().MongoDb

type Connection struct {
	Used   bool
	Pos    int
	Client *mongo.Client
}

func (c *Connection) TryConnection() (*mongo.Database, error) {
	err := c.Client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return c.Client.Database(cfg.DataBase), err
}

type ConnectionPool struct {
	ClientList []Connection
}

var CP = ConnectionPool{}

func GetClient() (*Connection, error) {
	for i := 0; i < len(CP.ClientList); i++ {
		if CP.ClientList[i].Used == false {
			CP.ClientList[i].Used = true
			return &CP.ClientList[i], nil
		}
	}
	pos := len(CP.ClientList)
	if pos >= MaxConnectionNum {
		return nil, errors.New("Mongo连接池连接量不足")
	}
	database, err := DbConnect()
	if err != nil {
		return nil, err
	}
	connection := Connection{
		Used:   false,
		Pos:    pos,
		Client: database,
	}

	CP.ClientList = append(CP.ClientList, connection)
	return &connection, nil
}
func Release(pos int) {
	CP.ClientList[pos].Used = false
}

func DbConnect() (*mongo.Client, error) {
	var param string
	if cfg.String != "" {
		param = cfg.String
	} else {
		param = fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Name, cfg.Pass, cfg.Host, cfg.Port)
	}

	clientOptions := options.Client().ApplyURI(param)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	return client, nil
}
