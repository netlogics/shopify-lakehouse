// Package producer publishes Avro-encoded Shopify product and inventory
// events to Kafka via Confluent Schema Registry, using the exact schemas in
// schemas/product.avsc and schemas/inventory_level.avsc as the source of
// truth (registered verbatim, not derived from Go struct reflection).
package producer

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	avro "github.com/hamba/avro/v2"

	"generator/internal/config"
	"generator/internal/model"
)

// Producer publishes product and inventory events to their configured
// Kafka topics, encoded per the Confluent wire format (magic byte + 4-byte
// schema ID + Avro binary).
type Producer struct {
	kafka *kafka.Producer
	sr    schemaregistry.Client

	productsTopic  string
	inventoryTopic string

	productSchema     avro.Schema
	productSchemaID   int
	inventorySchema   avro.Schema
	inventorySchemaID int
}

// New builds a Producer: it connects to Kafka and the Schema Registry,
// registers the two schemas loaded from schemasDir (product.avsc and
// inventory_level.avsc) under the standard "<topic>-value" subjects, and
// returns a ready-to-use Producer.
func New(cfg *config.Config, schemasDir string) (*Producer, error) {
	kp, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Brokers,
	})
	if err != nil {
		return nil, fmt.Errorf("creating kafka producer: %w", err)
	}

	srClient, err := schemaregistry.NewClient(schemaregistry.NewConfig(cfg.Kafka.SchemaRegistry))
	if err != nil {
		kp.Close()
		return nil, fmt.Errorf("creating schema registry client: %w", err)
	}

	productSchema, productSchemaID, err := loadAndRegister(
		srClient, filepath.Join(schemasDir, "product.avsc"), cfg.Products.Topic+"-value")
	if err != nil {
		kp.Close()
		return nil, err
	}

	inventorySchema, inventorySchemaID, err := loadAndRegister(
		srClient, filepath.Join(schemasDir, "inventory_level.avsc"), cfg.Inventory.Topic+"-value")
	if err != nil {
		kp.Close()
		return nil, err
	}

	return &Producer{
		kafka:             kp,
		sr:                srClient,
		productsTopic:     cfg.Products.Topic,
		inventoryTopic:    cfg.Inventory.Topic,
		productSchema:     productSchema,
		productSchemaID:   productSchemaID,
		inventorySchema:   inventorySchema,
		inventorySchemaID: inventorySchemaID,
	}, nil
}

func loadAndRegister(client schemaregistry.Client, path, subject string) (avro.Schema, int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("reading schema %q: %w", path, err)
	}

	schema, err := avro.Parse(string(raw))
	if err != nil {
		return nil, 0, fmt.Errorf("parsing schema %q: %w", path, err)
	}

	id, err := client.Register(subject, schemaregistry.SchemaInfo{Schema: string(raw)}, false)
	if err != nil {
		return nil, 0, fmt.Errorf("registering subject %q: %w", subject, err)
	}

	return schema, id, nil
}

// encode wraps Avro binary data in the Confluent wire format: a leading
// zero magic byte followed by the big-endian 4-byte schema ID.
func encode(schemaID int, avroBytes []byte) []byte {
	buf := make([]byte, 5+len(avroBytes))
	buf[0] = 0
	binary.BigEndian.PutUint32(buf[1:5], uint32(schemaID))
	copy(buf[5:], avroBytes)
	return buf
}

// PublishProduct encodes and produces a product event, keyed by product ID.
func (p *Producer) PublishProduct(product model.Product) error {
	avroBytes, err := avro.Marshal(p.productSchema, product)
	if err != nil {
		return fmt.Errorf("encoding product: %w", err)
	}

	topic := p.productsTopic
	key := strconv.FormatInt(product.ID, 10)
	return p.kafka.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          encode(p.productSchemaID, avroBytes),
	}, nil)
}

// PublishInventoryLevel encodes and produces an inventory event, keyed by SKU.
func (p *Producer) PublishInventoryLevel(level model.InventoryLevel) error {
	avroBytes, err := avro.Marshal(p.inventorySchema, level)
	if err != nil {
		return fmt.Errorf("encoding inventory level: %w", err)
	}

	topic := p.inventoryTopic
	key := level.SKU
	return p.kafka.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          encode(p.inventorySchemaID, avroBytes),
	}, nil)
}

// Events exposes the underlying producer's event channel (delivery reports
// and errors) for callers that want to log or track them.
func (p *Producer) Events() chan kafka.Event {
	return p.kafka.Events()
}

// Flush blocks until all outstanding messages are delivered or the timeout
// (in milliseconds) elapses, returning the number of messages still
// outstanding.
func (p *Producer) Flush(timeoutMs int) int {
	return p.kafka.Flush(timeoutMs)
}

// Close flushes outstanding messages and releases the Kafka producer and
// Schema Registry client.
func (p *Producer) Close() {
	p.kafka.Flush(10_000)
	p.kafka.Close()
	if p.sr != nil {
		_ = p.sr.Close()
	}
}
