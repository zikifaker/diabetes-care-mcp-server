package dao

import (
	"context"
	"diabetes-agent-mcp-server/config"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var Driver neo4j.DriverWithContext

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Driver.VerifyConnectivity(ctx); err != nil {
		panic(fmt.Sprintf("Failed to connect to Neo4j server: %v", err))
	}
}
