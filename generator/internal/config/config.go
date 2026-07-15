// Package config loads generator configuration from a YAML file with
// environment-variable overrides and sensible defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level generator configuration.
type Config struct {
	Kafka     KafkaConfig     `yaml:"kafka"`
	Products  ProductsConfig  `yaml:"products"`
	Inventory InventoryConfig `yaml:"inventory"`
}

// KafkaConfig holds broker and schema registry connection settings.
type KafkaConfig struct {
	Brokers        string `yaml:"brokers"`
	SchemaRegistry string `yaml:"schema_registry"`
}

// ProductsConfig controls product-creation event emission.
type ProductsConfig struct {
	Topic     string `yaml:"topic"`
	Rate      string `yaml:"rate"`
	SeedCount int    `yaml:"seed_count"`
}

// InventoryConfig controls inventory-update event emission.
type InventoryConfig struct {
	Topic     string `yaml:"topic"`
	Rate      string `yaml:"rate"`
	Locations int    `yaml:"locations"`
}

func defaults() Config {
	return Config{
		Kafka: KafkaConfig{
			Brokers:        "kafka:9092",
			SchemaRegistry: "http://schema-registry:8081",
		},
		Products: ProductsConfig{
			Topic:     "shopify.products",
			Rate:      "1/s",
			SeedCount: 100,
		},
		Inventory: InventoryConfig{
			Topic:     "shopify.inventory",
			Rate:      "10/s",
			Locations: 3,
		},
	}
}

// Load reads a YAML config file (if it exists), fills in defaults for any
// zero-valued fields, then applies environment-variable overrides.
func Load(path string) (*Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config file %q: %w", path, err)
			}
		} else {
			var fileCfg Config
			if err := yaml.Unmarshal(data, &fileCfg); err != nil {
				return nil, fmt.Errorf("parsing config file %q: %w", path, err)
			}
			applyNonZero(&cfg, &fileCfg)
		}
	}

	applyEnvOverrides(&cfg)

	if _, err := ParseRate(cfg.Products.Rate); err != nil {
		return nil, fmt.Errorf("products.rate: %w", err)
	}
	if _, err := ParseRate(cfg.Inventory.Rate); err != nil {
		return nil, fmt.Errorf("inventory.rate: %w", err)
	}

	return &cfg, nil
}

func applyNonZero(dst, src *Config) {
	if src.Kafka.Brokers != "" {
		dst.Kafka.Brokers = src.Kafka.Brokers
	}
	if src.Kafka.SchemaRegistry != "" {
		dst.Kafka.SchemaRegistry = src.Kafka.SchemaRegistry
	}
	if src.Products.Topic != "" {
		dst.Products.Topic = src.Products.Topic
	}
	if src.Products.Rate != "" {
		dst.Products.Rate = src.Products.Rate
	}
	if src.Products.SeedCount != 0 {
		dst.Products.SeedCount = src.Products.SeedCount
	}
	if src.Inventory.Topic != "" {
		dst.Inventory.Topic = src.Inventory.Topic
	}
	if src.Inventory.Rate != "" {
		dst.Inventory.Rate = src.Inventory.Rate
	}
	if src.Inventory.Locations != 0 {
		dst.Inventory.Locations = src.Inventory.Locations
	}
}

// applyEnvOverrides overlays environment variables on top of the config.
// KAFKA_BOOTSTRAP / SCHEMA_REGISTRY_URL match the shared root .env keys;
// the rest are generator-specific.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("KAFKA_BOOTSTRAP"); v != "" {
		cfg.Kafka.Brokers = v
	}
	if v := os.Getenv("SCHEMA_REGISTRY_URL"); v != "" {
		cfg.Kafka.SchemaRegistry = v
	}
	if v := os.Getenv("PRODUCTS_TOPIC"); v != "" {
		cfg.Products.Topic = v
	}
	if v := os.Getenv("PRODUCTS_RATE"); v != "" {
		cfg.Products.Rate = v
	}
	if v := os.Getenv("PRODUCTS_SEED_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Products.SeedCount = n
		}
	}
	if v := os.Getenv("INVENTORY_TOPIC"); v != "" {
		cfg.Inventory.Topic = v
	}
	if v := os.Getenv("INVENTORY_RATE"); v != "" {
		cfg.Inventory.Rate = v
	}
	if v := os.Getenv("INVENTORY_LOCATIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Inventory.Locations = n
		}
	}
}

// ParseRate parses rate strings of the form "<count>/<duration>", e.g.
// "1/s", "5/s", "1/10s", into the interval between successive events.
func ParseRate(rate string) (time.Duration, error) {
	parts := strings.SplitN(rate, "/", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rate %q: expected format like \"1/s\" or \"5/s\"", rate)
	}

	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || count <= 0 {
		return 0, fmt.Errorf("invalid rate %q: count must be a positive integer", rate)
	}

	durStr := strings.TrimSpace(parts[1])
	if durStr == "" {
		return 0, fmt.Errorf("invalid rate %q: missing duration", rate)
	}
	if durStr[0] < '0' || durStr[0] > '9' {
		durStr = "1" + durStr
	}

	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return 0, fmt.Errorf("invalid rate %q: %w", rate, err)
	}
	if dur <= 0 {
		return 0, fmt.Errorf("invalid rate %q: duration must be positive", rate)
	}

	return dur / time.Duration(count), nil
}
