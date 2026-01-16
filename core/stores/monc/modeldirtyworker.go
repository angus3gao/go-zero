package monc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	ModelPoolPackage interface {
		Get() any
		Put(any)
	}
)

type ModelDirtyWorker struct {
	mongo *mon.Database
	redis *redis.Redis

	modelPoolPackages map[string]ModelPoolPackage
}

func MustNewModelDirtyWorker(mongo *mon.Database, redis *redis.Redis) *ModelDirtyWorker {
	return &ModelDirtyWorker{
		mongo:             mongo,
		redis:             redis,
		modelPoolPackages: map[string]ModelPoolPackage{},
	}
}

func (mdw ModelDirtyWorker) AddModel(name string, modelFun ModelPoolPackage) {
	mdw.modelPoolPackages[name] = modelFun
}

func (mdw ModelDirtyWorker) DirtyWorker(ctx context.Context) {
	for {
		keys, err := mdw.redis.LrangeCtx(ctx, "dirty:queue", 0, 10)
		if err != nil {
			log.Println("LrangeCtx error:", err)
			continue
		}

		for _, key := range keys {
			if err := mdw.persist(ctx, key); err != nil {
				log.Println("persist failed:", err)
				continue
			}

			// 落地成功后移除脏标记
			mdw.redis.SremCtx(ctx, "dirty:set", key)
		}
	}
}

func (mdw ModelDirtyWorker) persist(ctx context.Context, key string) error {
	valStr, err := mdw.redis.GetCtx(ctx, key)
	if err != nil {
		return err
	}

	subKeys := strings.Split(key, ":")
	name := subKeys[2]
	modelPoolPackage, ok := mdw.modelPoolPackages[name]
	if !ok {
		return fmt.Errorf("model[%s] not found", name)
	}
	model := modelPoolPackage.Get()
	err = json.Unmarshal([]byte(valStr), model)
	if err != nil {
		return fmt.Errorf("model[%s] value json.Unmarshal error: %v", name, err)
	}

	_, err = mdw.mongo.Collection(name).UpdateOne(ctx, bson.M{subKeys[3]: subKeys[4]}, bson.M{"$set": model})
	modelPoolPackage.Put(model)
	return err
}
