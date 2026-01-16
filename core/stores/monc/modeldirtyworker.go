package monc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
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
	num := 0
	for {
		keys, err := mdw.redis.RpopCountCtx(ctx, "dirty:queue", 100)
		if err == redis.Nil {
			logx.Infof("DirtyWorker rem.num: %d", num)
			num = 0
			time.Sleep(10 * time.Second)
			continue
		}
		if err != nil {
			logx.Errorf("LrangeCtx error: %+v", err)
			continue
		}

		values, err := mdw.redis.MgetNoKeysPrefixCtx(ctx, keys...)
		if err != nil {
			logx.Errorf("MgetCtx error: %+v", err)
			continue
		}
		for idx, key := range keys {
			if err := mdw.persist(ctx, key, values[idx]); err != nil {
				logx.Errorf("persist failed: %+v", err)
				time.Sleep(30 * time.Second)
				continue
			}
		}

		// 落地成功后移除脏标记
		rnum, err := mdw.redis.SremCtx(ctx, "dirty:set", keys)
		num += rnum
		if err != nil {
			logx.Errorf("redis.SremCtx error: %+v", err)
			time.Sleep(30 * time.Second)
		}
	}
}

func (mdw ModelDirtyWorker) persist(ctx context.Context, key, valStr string) error {
	subKeys := strings.Split(key, ":")
	name := subKeys[2]
	modelPoolPackage, ok := mdw.modelPoolPackages[name]
	if !ok {
		return fmt.Errorf("model[%s] not found", name)
	}
	model := modelPoolPackage.Get()
	err := json.Unmarshal([]byte(valStr), model)

	if err != nil {
		return fmt.Errorf("model[%s] value json.Unmarshal error: %v", name, err)
	}

	_, err = mdw.mongo.Collection(name).UpdateOne(ctx, bson.M{subKeys[3]: subKeys[4]}, bson.M{"$set": model})
	modelPoolPackage.Put(model)
	return err
}
