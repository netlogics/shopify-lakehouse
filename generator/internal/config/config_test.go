package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseRate(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"1/s", time.Second, false},
		{"5/s", 200 * time.Millisecond, false},
		{"10/s", 100 * time.Millisecond, false},
		{"1/10s", 10 * time.Second, false},
		{"2/10s", 5 * time.Second, false},
		{"", 0, true},
		{"1", 0, true},
		{"0/s", 0, true},
		{"-1/s", 0, true},
		{"1/", 0, true},
		{"one/s", 0, true},
		{"1/bogus", 0, true},
	}

	for _, tc := range cases {
		got, err := ParseRate(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseRate(%q): expected error, got nil", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRate(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseRate(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\"): unexpected error: %v", err)
	}
	if cfg.Kafka.Brokers != "kafka:9092" {
		t.Errorf("Kafka.Brokers = %q, want %q", cfg.Kafka.Brokers, "kafka:9092")
	}
	if cfg.Kafka.SchemaRegistry != "http://schema-registry:8081" {
		t.Errorf("Kafka.SchemaRegistry = %q", cfg.Kafka.SchemaRegistry)
	}
	if cfg.Products.Topic != "shopify.products" {
		t.Errorf("Products.Topic = %q", cfg.Products.Topic)
	}
	if cfg.Products.Rate != "1/s" {
		t.Errorf("Products.Rate = %q", cfg.Products.Rate)
	}
	if cfg.Products.SeedCount != 100 {
		t.Errorf("Products.SeedCount = %d", cfg.Products.SeedCount)
	}
	if cfg.Inventory.Topic != "shopify.inventory" {
		t.Errorf("Inventory.Topic = %q", cfg.Inventory.Topic)
	}
	if cfg.Inventory.Rate != "10/s" {
		t.Errorf("Inventory.Rate = %q", cfg.Inventory.Rate)
	}
	if cfg.Inventory.Locations != 3 {
		t.Errorf("Inventory.Locations = %d", cfg.Inventory.Locations)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yamlContent := `
kafka:
  brokers: custom-broker:9092
products:
  topic: custom.products
  rate: 5/s
  seed_count: 42
inventory:
  locations: 7
`
	if err := os.WriteFile(path, []byte(yamlContent), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(%q): unexpected error: %v", path, err)
	}

	if cfg.Kafka.Brokers != "custom-broker:9092" {
		t.Errorf("Kafka.Brokers = %q", cfg.Kafka.Brokers)
	}
	// Unset in file, should fall back to default.
	if cfg.Kafka.SchemaRegistry != "http://schema-registry:8081" {
		t.Errorf("Kafka.SchemaRegistry = %q", cfg.Kafka.SchemaRegistry)
	}
	if cfg.Products.Topic != "custom.products" {
		t.Errorf("Products.Topic = %q", cfg.Products.Topic)
	}
	if cfg.Products.Rate != "5/s" {
		t.Errorf("Products.Rate = %q", cfg.Products.Rate)
	}
	if cfg.Products.SeedCount != 42 {
		t.Errorf("Products.SeedCount = %d", cfg.Products.SeedCount)
	}
	if cfg.Inventory.Locations != 7 {
		t.Errorf("Inventory.Locations = %d", cfg.Inventory.Locations)
	}
	// Unset in file, should fall back to default.
	if cfg.Inventory.Rate != "10/s" {
		t.Errorf("Inventory.Rate = %q", cfg.Inventory.Rate)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("KAFKA_BOOTSTRAP", "env-broker:9092")
	t.Setenv("SCHEMA_REGISTRY_URL", "http://env-sr:8081")
	t.Setenv("PRODUCTS_TOPIC", "env.products")
	t.Setenv("PRODUCTS_RATE", "2/s")
	t.Setenv("PRODUCTS_SEED_COUNT", "9")
	t.Setenv("INVENTORY_TOPIC", "env.inventory")
	t.Setenv("INVENTORY_RATE", "1/5s")
	t.Setenv("INVENTORY_LOCATIONS", "4")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\"): unexpected error: %v", err)
	}

	if cfg.Kafka.Brokers != "env-broker:9092" {
		t.Errorf("Kafka.Brokers = %q", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.SchemaRegistry != "http://env-sr:8081" {
		t.Errorf("Kafka.SchemaRegistry = %q", cfg.Kafka.SchemaRegistry)
	}
	if cfg.Products.Topic != "env.products" {
		t.Errorf("Products.Topic = %q", cfg.Products.Topic)
	}
	if cfg.Products.Rate != "2/s" {
		t.Errorf("Products.Rate = %q", cfg.Products.Rate)
	}
	if cfg.Products.SeedCount != 9 {
		t.Errorf("Products.SeedCount = %d", cfg.Products.SeedCount)
	}
	if cfg.Inventory.Topic != "env.inventory" {
		t.Errorf("Inventory.Topic = %q", cfg.Inventory.Topic)
	}
	if cfg.Inventory.Rate != "1/5s" {
		t.Errorf("Inventory.Rate = %q", cfg.Inventory.Rate)
	}
	if cfg.Inventory.Locations != 4 {
		t.Errorf("Inventory.Locations = %d", cfg.Inventory.Locations)
	}
}

func TestLoadInvalidRate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("products:\n  rate: bogus\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load: expected error for invalid rate, got nil")
	}
}
