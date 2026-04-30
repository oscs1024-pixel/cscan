package svc

import (
	"context"
	"fmt"
	"time"

	"cscan/model"
	"cscan/rpc/task/internal/config"
	"cscan/scheduler"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ServiceContext struct {
	Config                  config.Config
	MongoClient             *mongo.Client
	MongoDB                 *mongo.Database
	RedisClient             *redis.Client
	NucleiTemplateModel     *model.NucleiTemplateModel
	FingerprintModel        *model.FingerprintModel
	CustomPocModel          *model.CustomPocModel
	HttpServiceMappingModel *model.HttpServiceMappingModel
	WorkspaceModel          *model.WorkspaceModel
	SubfinderProviderModel  *model.SubfinderProviderModel
	NotifyConfigModel       *model.NotifyConfigModel
	TaskRecoveryManager     *scheduler.TaskRecoveryManager // 任务恢复管理器
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	logx.Infof("Connecting to MongoDB: %s", c.Mongo.Uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(c.Mongo.Uri))
	if err != nil {
		return nil, fmt.Errorf("connect MongoDB: %w", err)
	}

	// 测试 MongoDB 连接
	if err := mongoClient.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}
	logx.Info("MongoDB connected successfully")

	mongoDB := mongoClient.Database(c.Mongo.DbName)

	// 使用go-zero Redis配置
	logx.Infof("Connecting to Redis: %s", c.RedisConf.Host)
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConf.Host,
		Password: c.RedisConf.Pass,
		DB:       0,
	})

	// 测试 Redis 连接
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping Redis: %w", err)
	}
	logx.Info("Redis connected successfully")

	// 创建任务恢复管理器
	recoveryManager := scheduler.NewTaskRecoveryManager(rdb, context.Background())
	recoveryManager.Start()
	logx.Info("Task recovery manager started")

	return &ServiceContext{
		Config:                  c,
		MongoClient:             mongoClient,
		MongoDB:                 mongoDB,
		RedisClient:             rdb,
		NucleiTemplateModel:     model.NewNucleiTemplateModel(mongoDB),
		FingerprintModel:        model.NewFingerprintModel(mongoDB),
		CustomPocModel:          model.NewCustomPocModel(mongoDB),
		HttpServiceMappingModel: model.NewHttpServiceMappingModel(mongoDB),
		WorkspaceModel:          model.NewWorkspaceModel(mongoDB),
		SubfinderProviderModel:  model.NewSubfinderProviderModel(mongoDB),
		NotifyConfigModel:       model.NewNotifyConfigModel(mongoDB),
		TaskRecoveryManager:     recoveryManager,
	}, nil
}

func (s *ServiceContext) GetAssetModel(workspaceId string) *model.AssetModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetMainTaskModel(workspaceId string) *model.MainTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewMainTaskModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetVulModel(workspaceId string) *model.VulModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewVulModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetExecutorTaskModel(workspaceId string) *model.ExecutorTaskModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewExecutorTaskModel(s.MongoDB, workspaceId)
}

func (s *ServiceContext) GetAssetHistoryModel(workspaceId string) *model.AssetHistoryModel {
	if workspaceId == "" {
		workspaceId = "default"
	}
	return model.NewAssetHistoryModel(s.MongoDB, workspaceId)
}
