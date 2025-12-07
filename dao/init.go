package dao

import (
	"context"
	"diabetes-agent-mcp-server/config"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	Driver       neo4j.DriverWithContext
	MilvusClient *milvusclient.Client
)

// 初始化 Neo4j 连接
func init() {
	dsn := fmt.Sprintf("neo4j://%s:%s", config.Cfg.DB.Neo4j.Host, config.Cfg.DB.Neo4j.Port)

	var err error
	Driver, err = neo4j.NewDriverWithContext(
		dsn,
		neo4j.BasicAuth(config.Cfg.DB.Neo4j.Username, config.Cfg.DB.Neo4j.Password, ""),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Neo4j driver: %v", err))
	}

	ctx := context.Background()
	if err := Driver.VerifyConnectivity(ctx); err != nil {
		panic(fmt.Sprintf("Failed to connect to Neo4j server: %v", err))
	}
}

// 初始化 Milvus 客户端
func init() {
	milvusConfig := milvusclient.ClientConfig{
		Address: config.Cfg.Milvus.Endpoint,
		APIKey:  config.Cfg.Milvus.APIKey,
	}

	var err error
	MilvusClient, err = milvusclient.New(context.Background(), &milvusConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Milvus client: %v", err))
	}
}
