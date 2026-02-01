package dao

import (
	"context"
	"diabetes-care-mcp-server/config"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Driver neo4j.DriverWithContext
	DB     *gorm.DB
)

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

func init() {
	dbConfig := config.Cfg.DB.MySQL

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(fmt.Sprintf("Failed to connect database: %v", err))
	}
}
