package main

import (
	"context"
	"diabetes-care-mcp-server/config"
	"diabetes-care-mcp-server/tools"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const createIndexTimeout = 10

type Entity struct {
	EntityID   string `json:"entity_id"`
	Entity     string `json:"entity"`
	EntityType string `json:"entity_type"`
	StartIdx   int    `json:"start_idx"`
	EndIdx     int    `json:"end_idx"`
}

type Relation struct {
	RelationType string `json:"relation_type"`
	RelationID   string `json:"relation_id"`
	HeadEntityID string `json:"head_entity_id"`
	TailEntityID string `json:"tail_entity_id"`
}

type Sentence struct {
	SentenceID string     `json:"sentence_id"`
	Sentence   string     `json:"sentence"`
	StartIdx   int        `json:"start_idx"`
	EndIdx     int        `json:"end_idx"`
	Entities   []Entity   `json:"entities"`
	Relations  []Relation `json:"relations"`
}

type Paragraph struct {
	ParagraphID string     `json:"paragraph_id"`
	Paragraph   string     `json:"paragraph"`
	Sentences   []Sentence `json:"sentences"`
}

type Document struct {
	DocID      string      `json:"doc_id"`
	Paragraphs []Paragraph `json:"paragraphs"`
}

func main() {
	ctx := context.Background()
	dsn := fmt.Sprintf("neo4j://%s:%s", config.Cfg.DB.Neo4j.Host, config.Cfg.DB.Neo4j.Port)

	driver, err := neo4j.NewDriverWithContext(
		dsn,
		neo4j.BasicAuth(config.Cfg.DB.Neo4j.Username, config.Cfg.DB.Neo4j.Password, ""),
	)
	if err != nil {
		slog.Error("Failed to create Neo4j driver", "err", err)
		return
	}
	defer driver.Close(ctx)

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		slog.Error("Failed to connect to Neo4j", "err", err)
		return
	}

	// 读取数据集
	files, err := filepath.Glob("resource/diakg/*.json")
	if err != nil {
		slog.Error("Error reading directory", "err", err)
		return
	}

	for _, file := range files {
		slog.Info("Processing file", "file", file)
		err = processFile(ctx, driver, file)
		if err != nil {
			slog.Error("Error processing file",
				"file", file,
				"err", err,
			)
			return
		}
	}

	// 检查全文索引，若不存在进行创建
	if err := checkFullTextIndex(ctx, driver); err != nil {
		slog.Error("Failed to check fulltext index", "err", err)
		return
	}

	slog.Info("Created knowledge graph successfully")
}

func processFile(ctx context.Context, driver neo4j.DriverWithContext, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return saveDocument(ctx, driver, doc)
}

func saveDocument(ctx context.Context, driver neo4j.DriverWithContext, doc Document) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		if err := saveDocumentNode(tx, doc); err != nil {
			return nil, fmt.Errorf("error saving document node %s: %v", doc.DocID, err)
		}

		for _, para := range doc.Paragraphs {
			if err := saveParagraphNode(tx, doc.DocID, para); err != nil {
				return nil, fmt.Errorf("error saving paragraph node %s: %v", para.ParagraphID, err)
			}
		}

		return nil, nil
	})

	return err
}

func saveDocumentNode(tx neo4j.ManagedTransaction, doc Document) error {
	query := `
		MERGE (d:Document {doc_id: $doc_id})
		RETURN d
    `
	_, err := tx.Run(context.Background(), query, map[string]any{
		"doc_id": doc.DocID,
	})
	return err
}

func saveParagraphNode(tx neo4j.ManagedTransaction, docID string, para Paragraph) error {
	query := `
		MATCH (d:Document {doc_id: $doc_id})
		MERGE (p:Paragraph {paragraph_id: $para_id})
		SET p.text = $text
		MERGE (d)-[:CONTAINS_PARAGRAPH]->(p)
    `
	if _, err := tx.Run(context.Background(), query, map[string]any{
		"doc_id":  docID,
		"para_id": para.ParagraphID,
		"text":    para.Paragraph,
	}); err != nil {
		return fmt.Errorf("error creating paragraph node: %v", err)
	}

	for _, sentence := range para.Sentences {
		if err := saveSentenceNode(tx, para.ParagraphID, sentence); err != nil {
			return fmt.Errorf("error saving sentence node %s: %v", sentence.SentenceID, err)
		}
	}

	return nil
}

func saveSentenceNode(tx neo4j.ManagedTransaction, paraID string, sentence Sentence) error {
	query := `
		MATCH (p:Paragraph {paragraph_id: $para_id})
		MERGE (s:Sentence {sentence_id: $sentence_id})
		SET s.text = $text
		MERGE (p)-[:CONTAINS_SENTENCE]->(s)
    `
	if _, err := tx.Run(context.Background(), query, map[string]any{
		"para_id":     paraID,
		"sentence_id": sentence.SentenceID,
		"text":        sentence.Sentence,
	}); err != nil {
		return fmt.Errorf("error creating sentence node: %v", err)
	}

	for _, entity := range sentence.Entities {
		if err := saveEntityNode(tx, sentence.SentenceID, entity); err != nil {
			return fmt.Errorf("error saving entity node %s: %v", entity.EntityID, err)
		}
	}

	for _, relation := range sentence.Relations {
		if err := saveRelation(tx, relation); err != nil {
			return fmt.Errorf("error saving relation %s: %v", relation.RelationID, err)
		}
	}

	return nil
}

func saveEntityNode(tx neo4j.ManagedTransaction, sentenceID string, entity Entity) error {
	query := `
		MATCH (s:Sentence {sentence_id: $sentence_id})
		MERGE (e:Entity {entity_id: $entity_id})
		SET e.name = $name,
			e.type = $type
		MERGE (s)-[:CONTAINS_ENTITY]->(e)
    `
	_, err := tx.Run(context.Background(), query, map[string]any{
		"sentence_id": sentenceID,
		"entity_id":   entity.EntityID,
		"name":        entity.Entity,
		"type":        entity.EntityType,
	})
	if err != nil {
		return err
	}
	return nil
}

func saveRelation(tx neo4j.ManagedTransaction, relation Relation) error {
	query := fmt.Sprintf(`
		MATCH (e1:Entity {entity_id: $head_id})
		MATCH (e2:Entity {entity_id: $tail_id})
		MERGE (e1)-[r:%s]->(e2)
		SET r.relation_id = $relation_id
	`, relation.RelationType)

	_, err := tx.Run(context.Background(), query, map[string]interface{}{
		"head_id":     relation.HeadEntityID,
		"tail_id":     relation.TailEntityID,
		"relation_id": relation.RelationID,
	})
	if err != nil {
		return err
	}
	return nil
}

func checkFullTextIndex(ctx context.Context, driver neo4j.DriverWithContext) error {
	s := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	})
	defer s.Close(ctx)

	check := `
		SHOW FULLTEXT INDEXES
		YIELD name
		WHERE name = $name
		RETURN count(*) AS count
	`
	res, err := s.Run(ctx, check, map[string]any{"name": tools.Neo4jFulltextIndexName})
	if err != nil {
		return fmt.Errorf("failed to list fulltext indexes: %w", err)
	}

	idxExists := false
	if res.Next(ctx) {
		if cnt, ok := res.Record().Get("count"); ok {
			if c, ok2 := cnt.(int64); ok2 && c > 0 {
				idxExists = true
			}
		}
	}
	if err := res.Err(); err != nil {
		return fmt.Errorf("failed to read index list: %v", err)
	}
	if idxExists {
		slog.Info(fmt.Sprintf("%s index already exists", tools.Neo4jFulltextIndexName))
		return nil
	}

	// 在 Entity 节点的 name 属性上建立全文索引
	create := fmt.Sprintf(`CREATE FULLTEXT INDEX %s FOR (n:Entity) ON EACH [n.name]`, tools.Neo4jFulltextIndexName)

	_, err = s.Run(ctx, create, nil)
	if err != nil {
		return fmt.Errorf("failed to create fulltext index: %v", err)
	}

	// 等待全文索引创建完成
	wait := fmt.Sprintf(`CALL db.awaitIndexes(%d)`, createIndexTimeout)
	_, err = s.Run(ctx, wait, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for index creation: %v", err)
	}

	slog.Info(fmt.Sprintf("Successfully created index: %s", tools.Neo4jFulltextIndexName))

	return nil
}
