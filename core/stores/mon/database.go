package mon

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

type (
	Database struct {
		name     string
		database *mongo.Database
		brk      breaker.Breaker
		opts     []Option
	}
)

func MustNewDatabase(uri, db, collection string, opts ...Option) *Database {
	database, err := NewDatabase(uri, db, opts...)
	logx.Must(err)
	return database
}

func NewDatabase(uri, db string, opts ...Option) (*Database, error) {
	cli, err := getClient(uri, opts...)
	if err != nil {
		return nil, err
	}

	name := strings.Join([]string{uri, db}, "/")
	brk := breaker.GetBreaker(uri)
	return newDatabase(name, cli.Database(db), brk, opts...), nil
}

func newDatabase(name string, database *mongo.Database, brk breaker.Breaker,
	opts ...Option) *Database {
	return &Database{
		name:     name,
		database: database,
		brk:      brk,
		opts:     opts,
	}
}

func (db *Database) Collection(cllection string) Collection {
	return newCollection(db.database.Collection(cllection), db.brk)
}

func (db *Database) CreateIndex(ctx context.Context, collection string, unique bool, fields ...string) error {
	keys := bson.D{}
	for _, field := range fields {
		keys = append(keys, bson.E{Key: field, Value: 1})
	}
	_, err := db.Collection(collection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: mongoOptions.Index().SetUnique(unique),
	})
	if err != nil {
		if strings.Contains(err.Error(), "with different options") {
			return nil
		}
		return err
	}
	return nil
}
