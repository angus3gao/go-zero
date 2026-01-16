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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	ModelPoolPackage interface {
		Get() any
		Put(any)
	}

	ModelDirtyWorkerConf struct {
		ErrorWait int `json:",default=10"`  // 错误等待时间
		NilWait   int `json:",default=30"`  // 空等待时间
		OneNum    int `json:",default=100"` // 单次处理数量
	}
)

type ModelDirtyWorker struct {
	mongo *mon.Database
	redis *redis.Redis

	modelPoolPackages map[string]ModelPoolPackage
	conf              ModelDirtyWorkerConf
}

func MustNewModelDirtyWorker(mongo *mon.Database, redis *redis.Redis, conf ModelDirtyWorkerConf) *ModelDirtyWorker {
	if conf.ErrorWait <= 0 {
		conf.ErrorWait = 10
	}
	if conf.NilWait <= 0 {
		conf.NilWait = 30
	}

	if conf.OneNum <= 0 {
		conf.OneNum = 100
	}

	return &ModelDirtyWorker{
		mongo:             mongo,
		redis:             redis,
		modelPoolPackages: map[string]ModelPoolPackage{},
		conf:              conf,
	}
}

func (mdw ModelDirtyWorker) AddModel(name string, modelFun ModelPoolPackage) {
	mdw.modelPoolPackages[name] = modelFun
}

func (mdw ModelDirtyWorker) DirtyWorker(ctx context.Context) {
	num := 0
	start := time.Now().Second()
	for {
		keys, err := mdw.redis.RpopCountCtx(ctx, "dirty:queue", mdw.conf.OneNum)
		if err == redis.Nil {
			if num > 0 {
				logx.Infof("DirtyWorker rem.num: %d", num)
			}

			num = 0
			less := mdw.conf.NilWait - (time.Now().Second() - start)
			time.Sleep(time.Duration(less) * time.Second)
			start = time.Now().Second()
			continue
		}

		if err != nil {
			logx.Errorf("LrangeCtx error: %+v", err)
			time.Sleep(time.Duration(mdw.conf.ErrorWait) * time.Second)
			continue
		}

		values, err := mdw.redis.MgetNoKeysPrefixCtx(ctx, keys...)
		if err != nil {
			logx.Errorf("MgetCtx error: %+v", err)
			mdw.redis.LpushCtx(ctx, "dirty:queue", keys)
			time.Sleep(time.Duration(mdw.conf.ErrorWait) * time.Second)
			continue
		}
		updateMap := map[string][]mongo.WriteModel{}
		for idx, key := range keys {
			name, update, err := mdw.persist(key, values[idx])
			if err != nil {
				logx.Errorf("persist value: %s failed: %+v", err, values[idx])
				continue
			}
			if updateList, ok := updateMap[name]; ok {
				updateMap[name] = append(updateList, update)
			} else {
				updateList = []mongo.WriteModel{}
				updateMap[name] = append(updateList, update)
			}
		}

		for name, updates := range updateMap {
			opts := options.BulkWrite().SetOrdered(false)
			_, err := mdw.mongo.Collection(name).BulkWrite(ctx, updates, opts)
			if err != nil {
				logx.Errorf("redis.SremCtx error: %+v", err)
				mdw.redis.LpushCtx(ctx, "dirty:queue", keys)
				time.Sleep(time.Duration(mdw.conf.ErrorWait) * time.Second)
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

func (mdw ModelDirtyWorker) persist(key, valStr string) (string, *mongo.UpdateOneModel, error) {
	subKeys := strings.Split(key, ":")
	name := subKeys[2]
	modelPoolPackage, ok := mdw.modelPoolPackages[name]
	if !ok {
		return "", nil, fmt.Errorf("model[%s] not found", name)
	}
	model := modelPoolPackage.Get()
	err := json.Unmarshal([]byte(valStr), model)

	if err != nil {
		return "", nil, fmt.Errorf("model[%s] value json.Unmarshal error: %v", name, err)
	}

	return name, mongo.NewUpdateOneModel().
		SetFilter(bson.M{subKeys[3]: subKeys[4]}).
		SetUpdate(bson.M{"$set": model}), nil
}
