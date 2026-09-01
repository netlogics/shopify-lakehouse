// Command generator seeds fake Shopify products and continuously emits new
// products, inventory-level updates, and order details to Kafka, Avro-encoded
// via Schema Registry.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"generator/internal/config"
	"generator/internal/gen"
	"generator/internal/metrics"
	"generator/internal/model"
	"generator/internal/producer"
)

// errorTracker tracks delivery errors in a sliding window to drive backpressure.
type errorTracker struct {
	mu        sync.Mutex
	errors    []time.Time
	window    time.Duration
	threshold int
}

func newErrorTracker(window time.Duration, threshold int) *errorTracker {
	return &errorTracker{window: window, threshold: threshold}
}

func (et *errorTracker) add() {
	et.mu.Lock()
	defer et.mu.Unlock()
	now := time.Now()
	et.errors = append(et.errors, now)
	// Trim entries outside the window.
	cutoff := now.Add(-et.window)
	i := 0
	for i < len(et.errors) && et.errors[i].Before(cutoff) {
		i++
	}
	if i > 0 {
		et.errors = et.errors[i:]
	}
}

func (et *errorTracker) shouldPause() bool {
	et.mu.Lock()
	defer et.mu.Unlock()
	return len(et.errors) >= et.threshold
}

func (et *errorTracker) count() int {
	et.mu.Lock()
	defer et.mu.Unlock()
	return len(et.errors)
}

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
	metricsAddr := os.Getenv("GENERATOR_METRICS_ADDR")
	if metricsAddr == "" {
		metricsAddr = ":2112"
	}
	go metrics.Serve(metricsAddr)

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
	orderDetailsInterval, err := config.ParseRate(cfg.OrderDetails.Rate)
	if err != nil {
		slog.Error("parsing order_details rate", "error", err)
		os.Exit(1)
	}
	customersInterval, err := config.ParseRate(cfg.Customers.Rate)
	if err != nil {
		slog.Error("parsing customers rate", "error", err)
		os.Exit(1)
	}
	fraudEpisodeDuration, err := time.ParseDuration(cfg.Fraud.EpisodeDuration)
	if err != nil {
		slog.Error("parsing fraud episode duration", "error", err)
		os.Exit(1)
	}
	fraudParams := gen.FraudParams{
		InjectionProbability: cfg.Fraud.InjectionProbability,
		EpisodeDuration:      fraudEpisodeDuration,
		TargetWeight:         cfg.Fraud.TargetWeight,
	}

	prod, err := producer.New(cfg, schemasDir)
	if err != nil {
		slog.Error("creating producer", "error", err)
		os.Exit(1)
	}
	defer prod.Close()

	// Backpressure config: pause publishing if >=5 errors in last 10 seconds.
	errorTracker := newErrorTracker(10*time.Second, 5)

	var productsSent, inventorySent, orderDetailsSent, customersSent, deliveryErrors atomic.Int64
	go logDeliveryEvents(prod.Events(), &deliveryErrors, errorTracker)

	registry := gen.NewRegistry()
	generator := gen.NewGenerator(gofakeit.New(0), registry)

	// variantMap keeps inventory_item_id → variant for O(1) lookups in
	// NewOrderDetail. It's bounded to prevent unbounded memory growth.
	const maxVariantMapSize = 5000
	variantMap := make(map[int64]*model.Variant, cfg.Products.SeedCount)

	slog.Info("seeding products", "count", cfg.Products.SeedCount)
	for i := 0; i < cfg.Products.SeedCount; i++ {
		p := generator.NewProduct()
		if err := prod.PublishProduct(p); err != nil {
			slog.Error("publishing seed product", "error", err)
			metrics.DeliveryErrors.WithLabelValues(cfg.Products.Topic).Inc()
			continue
		}
		for j := range p.Variants {
			v := &p.Variants[j]
			variantMap[v.InventoryItemID] = v
		}
		productsSent.Add(1)
		metrics.EventsProduced.WithLabelValues(cfg.Products.Topic).Inc()
	}
	prod.Flush(10_000)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	productsTicker := time.NewTicker(productsInterval)
	defer productsTicker.Stop()
	inventoryTicker := time.NewTicker(inventoryInterval)
	defer inventoryTicker.Stop()
	orderDetailsTicker := time.NewTicker(orderDetailsInterval)
	defer orderDetailsTicker.Stop()
	customersTicker := time.NewTicker(customersInterval)
	defer customersTicker.Stop()

	statsTicker := time.NewTicker(10 * time.Second)
	defer statsTicker.Stop()

	slog.Info("generator running",
		"products_topic", cfg.Products.Topic, "products_rate", cfg.Products.Rate,
		"inventory_topic", cfg.Inventory.Topic, "inventory_rate", cfg.Inventory.Rate,
		"order_details_topic", cfg.OrderDetails.Topic, "order_details_rate", cfg.OrderDetails.Rate,
		"customers_topic", cfg.Customers.Topic, "customers_rate", cfg.Customers.Rate,
	)

	// pausedUntil tracks when backpressure cooldown ends.
	var pausedUntil time.Time

	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down, flushing outstanding messages")
			prod.Flush(10_000)
			slog.Info("shutdown complete",
				"products_sent", productsSent.Load(),
				"inventory_sent", inventorySent.Load(),
				"order_details_sent", orderDetailsSent.Load(),
				"customers_sent", customersSent.Load(),
				"delivery_errors", deliveryErrors.Load(),
			)
			return

		case <-productsTicker.C:
			if pausedUntil.IsZero() && errorTracker.shouldPause() {
				pausedUntil = time.Now().Add(5 * time.Second)
				slog.Warn("backpressure: pausing due to delivery errors", "errors_in_window", errorTracker.count())
				metrics.BackpressurePauses.Inc()
				metrics.BackpressureActive.Set(1)
				continue
			}
			if !pausedUntil.IsZero() && time.Now().Before(pausedUntil) {
				continue
			}
			if !pausedUntil.IsZero() {
				pausedUntil = time.Time{}
				slog.Info("backpressure: resuming after cooldown")
				metrics.BackpressureActive.Set(0)
			}

			p := generator.NewProduct()
			if err := prod.PublishProduct(p); err != nil {
				slog.Error("publishing product", "error", err)
				metrics.DeliveryErrors.WithLabelValues(cfg.Products.Topic).Inc()
				continue
			}
			for j := range p.Variants {
				v := &p.Variants[j]
				variantMap[v.InventoryItemID] = v
			}
			productsSent.Add(1)
			metrics.EventsProduced.WithLabelValues(cfg.Products.Topic).Inc()
			// Prune map if it exceeds the limit (remove random entries to keep bounded).
			if len(variantMap) > maxVariantMapSize {
				toRemove := len(variantMap) - maxVariantMapSize
				i := 0
				for id := range variantMap {
					if i >= toRemove {
						break
					}
					delete(variantMap, id)
					i++
				}
			}

		case <-inventoryTicker.C:
			if pausedUntil.IsZero() && errorTracker.shouldPause() {
				pausedUntil = time.Now().Add(5 * time.Second)
				slog.Warn("backpressure: pausing due to delivery errors", "errors_in_window", errorTracker.count())
				metrics.BackpressurePauses.Inc()
				metrics.BackpressureActive.Set(1)
				continue
			}
			if !pausedUntil.IsZero() && time.Now().Before(pausedUntil) {
				continue
			}
			if !pausedUntil.IsZero() {
				pausedUntil = time.Time{}
				slog.Info("backpressure: resuming after cooldown")
				metrics.BackpressureActive.Set(0)
			}

			level, ok := generator.NewInventoryLevel(cfg.Inventory.Locations)
			if !ok {
				continue
			}
			if err := prod.PublishInventoryLevel(level); err != nil {
				slog.Error("publishing inventory level", "error", err)
				metrics.DeliveryErrors.WithLabelValues(cfg.Inventory.Topic).Inc()
				continue
			}
			inventorySent.Add(1)
			metrics.EventsProduced.WithLabelValues(cfg.Inventory.Topic).Inc()

		case <-orderDetailsTicker.C:
			if pausedUntil.IsZero() && errorTracker.shouldPause() {
				pausedUntil = time.Now().Add(5 * time.Second)
				slog.Warn("backpressure: pausing due to delivery errors", "errors_in_window", errorTracker.count())
				metrics.BackpressurePauses.Inc()
				metrics.BackpressureActive.Set(1)
				continue
			}
			if !pausedUntil.IsZero() && time.Now().Before(pausedUntil) {
				continue
			}
			if !pausedUntil.IsZero() {
				pausedUntil = time.Time{}
				slog.Info("backpressure: resuming after cooldown")
				metrics.BackpressureActive.Set(0)
			}

			detail, ok := generator.NewOrderDetail(variantMap, fraudParams)
			if !ok {
				continue
			}
			if err := prod.PublishOrderDetail(detail); err != nil {
				slog.Error("publishing order detail", "error", err)
				metrics.DeliveryErrors.WithLabelValues(cfg.OrderDetails.Topic).Inc()
				continue
			}
			orderDetailsSent.Add(1)
			metrics.EventsProduced.WithLabelValues(cfg.OrderDetails.Topic).Inc()

		case <-customersTicker.C:
			if pausedUntil.IsZero() && errorTracker.shouldPause() {
				pausedUntil = time.Now().Add(5 * time.Second)
				slog.Warn("backpressure: pausing due to delivery errors", "errors_in_window", errorTracker.count())
				metrics.BackpressurePauses.Inc()
				metrics.BackpressureActive.Set(1)
				continue
			}
			if !pausedUntil.IsZero() && time.Now().Before(pausedUntil) {
				continue
			}
			if !pausedUntil.IsZero() {
				pausedUntil = time.Time{}
				slog.Info("backpressure: resuming after cooldown")
				metrics.BackpressureActive.Set(0)
			}

			customer := generator.NewCustomer()
			if err := prod.PublishCustomer(customer); err != nil {
				slog.Error("publishing customer", "error", err)
				metrics.DeliveryErrors.WithLabelValues(cfg.Customers.Topic).Inc()
				continue
			}
			customersSent.Add(1)
			metrics.EventsProduced.WithLabelValues(cfg.Customers.Topic).Inc()

		case <-statsTicker.C:
			slog.Info("emit counts",
				"products_sent", productsSent.Load(),
				"inventory_sent", inventorySent.Load(),
				"order_details_sent", orderDetailsSent.Load(),
				"customers_sent", customersSent.Load(),
				"delivery_errors", deliveryErrors.Load(),
			)
		}
	}
}

// logDeliveryEvents drains the producer's event channel, logging delivery
// errors and counting successful deliveries per topic.
func logDeliveryEvents(events chan kafka.Event, deliveryErrors *atomic.Int64, errorTracker *errorTracker) {
	for e := range events {
		msg, ok := e.(*kafka.Message)
		if !ok {
			continue
		}
		if msg.TopicPartition.Error != nil {
			deliveryErrors.Add(1)
			errorTracker.add()
			slog.Error("delivery failed", "error", msg.TopicPartition.Error, "topic", *msg.TopicPartition.Topic)
			metrics.DeliveryErrors.WithLabelValues(*msg.TopicPartition.Topic).Inc()
		}
	}
}
