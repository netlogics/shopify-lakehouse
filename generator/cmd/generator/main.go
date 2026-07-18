// Command generator seeds fake Shopify products and continuously emits new
// products and inventory-level updates to Kafka, Avro-encoded via Schema
// Registry.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"generator/internal/config"
	"generator/internal/gen"
	"generator/internal/producer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	configPath := os.Getenv("GENERATOR_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	schemasDir := os.Getenv("GENERATOR_SCHEMAS_DIR")
	if schemasDir == "" {
		schemasDir = "schemas"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	productsInterval, err := config.ParseRate(cfg.Products.Rate)
	if err != nil {
		slog.Error("parsing products rate", "error", err)
		os.Exit(1)
	}
	inventoryInterval, err := config.ParseRate(cfg.Inventory.Rate)
	if err != nil {
		slog.Error("parsing inventory rate", "error", err)
		os.Exit(1)
	}

	prod, err := producer.New(cfg, schemasDir)
	if err != nil {
		slog.Error("creating producer", "error", err)
		os.Exit(1)
	}
	defer prod.Close()

	var productsSent, inventorySent, deliveryErrors atomic.Int64
	go logDeliveryEvents(prod.Events(), &deliveryErrors)

	registry := gen.NewRegistry()
	generator := gen.NewGenerator(gofakeit.New(0), registry)

	slog.Info("seeding products", "count", cfg.Products.SeedCount)
	for i := 0; i < cfg.Products.SeedCount; i++ {
		p := generator.NewProduct()
		if err := prod.PublishProduct(p); err != nil {
			slog.Error("publishing seed product", "error", err)
			continue
		}
		productsSent.Add(1)
	}
	prod.Flush(10_000)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	productsTicker := time.NewTicker(productsInterval)
	defer productsTicker.Stop()
	inventoryTicker := time.NewTicker(inventoryInterval)
	defer inventoryTicker.Stop()

	statsTicker := time.NewTicker(10 * time.Second)
	defer statsTicker.Stop()

	slog.Info("generator running",
		"products_topic", cfg.Products.Topic, "products_rate", cfg.Products.Rate,
		"inventory_topic", cfg.Inventory.Topic, "inventory_rate", cfg.Inventory.Rate,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down, flushing outstanding messages")
			prod.Flush(10_000)
			slog.Info("shutdown complete",
				"products_sent", productsSent.Load(),
				"inventory_sent", inventorySent.Load(),
				"delivery_errors", deliveryErrors.Load(),
			)
			return

		case <-productsTicker.C:
			p := generator.NewProduct()
			if err := prod.PublishProduct(p); err != nil {
				slog.Error("publishing product", "error", err)
				continue
			}
			productsSent.Add(1)

		case <-inventoryTicker.C:
			level, ok := generator.NewInventoryLevel(cfg.Inventory.Locations)
			if !ok {
				continue
			}
			if err := prod.PublishInventoryLevel(level); err != nil {
				slog.Error("publishing inventory level", "error", err)
				continue
			}
			inventorySent.Add(1)

		case <-statsTicker.C:
			slog.Info("emit counts",
				"products_sent", productsSent.Load(),
				"inventory_sent", inventorySent.Load(),
				"delivery_errors", deliveryErrors.Load(),
			)
		}
	}
}

// logDeliveryEvents drains the producer's event channel, logging delivery
// errors and counting successful deliveries per topic.
func logDeliveryEvents(events chan kafka.Event, deliveryErrors *atomic.Int64) {
	for e := range events {
		msg, ok := e.(*kafka.Message)
		if !ok {
			continue
		}
		if msg.TopicPartition.Error != nil {
			deliveryErrors.Add(1)
			slog.Error("delivery failed", "error", msg.TopicPartition.Error, "topic", *msg.TopicPartition.Topic)
		}
	}
}
