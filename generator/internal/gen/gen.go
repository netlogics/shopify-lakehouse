// Package gen produces fake Shopify product and inventory events, keeping an
// in-memory registry of known variants so inventory events always reference
// a real product/variant.
package gen

import (
	"fmt"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"generator/internal/model"
)

// VariantRef is the subset of variant data needed to emit a referentially
// valid inventory event.
type VariantRef struct {
	InventoryItemID int64
	SKU             string
	ProductID       int64
}

// Registry tracks every variant seen so far, used to pick a valid variant
// when generating inventory events.
type Registry struct {
	mu                  sync.Mutex
	variants            []VariantRef
	nextProductID       int64
	nextVariantID       int64
	nextInventoryItemID int64
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Len reports how many variants are currently known.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.variants)
}

func (r *Registry) addVariants(vs []VariantRef) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.variants = append(r.variants, vs...)
}

// RandomVariant picks a uniformly random known variant. ok is false if the
// registry is empty.
func (r *Registry) RandomVariant(f *gofakeit.Faker) (ref VariantRef, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.variants) == 0 {
		return VariantRef{}, false
	}
	idx := f.IntRange(0, len(r.variants)-1)
	return r.variants[idx], true
}

func (r *Registry) nextIDs(n int) (productID int64, variantIDs, inventoryItemIDs []int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextProductID++
	productID = r.nextProductID
	variantIDs = make([]int64, n)
	inventoryItemIDs = make([]int64, n)
	for i := 0; i < n; i++ {
		r.nextVariantID++
		variantIDs[i] = r.nextVariantID
		r.nextInventoryItemID++
		inventoryItemIDs[i] = r.nextInventoryItemID
	}
	return productID, variantIDs, inventoryItemIDs
}

// Generator produces fake product and inventory events. Faker is injected so
// tests can seed it for deterministic output.
type Generator struct {
	Faker    *gofakeit.Faker
	Registry *Registry
}

// NewGenerator builds a Generator backed by the given faker and registry.
func NewGenerator(f *gofakeit.Faker, reg *Registry) *Generator {
	return &Generator{Faker: f, Registry: reg}
}

// NewProduct creates a fake product with 1-3 variants, registering the new
// variants so later inventory events can reference them.
func (g *Generator) NewProduct() model.Product {
	f := g.Faker
	numVariants := f.IntRange(1, 3)
	productID, variantIDs, inventoryItemIDs := g.Registry.nextIDs(numVariants)

	variants := make([]model.Variant, numVariants)
	refs := make([]VariantRef, numVariants)
	for i := 0; i < numVariants; i++ {
		sku := fmt.Sprintf("SKU-%d-%d", productID, variantIDs[i])
		price := fmt.Sprintf("%.2f", f.Price(5, 500))
		variants[i] = model.Variant{
			ID:              variantIDs[i],
			SKU:             sku,
			Price:           price,
			InventoryItemID: inventoryItemIDs[i],
		}
		refs[i] = VariantRef{
			InventoryItemID: inventoryItemIDs[i],
			SKU:             sku,
			ProductID:       productID,
		}
	}
	g.Registry.addVariants(refs)

	numTags := f.IntRange(1, 4)
	tags := make([]string, numTags)
	for i := range tags {
		tags[i] = f.LoremIpsumWord()
	}

	now := time.Now().UnixMilli()
	return model.Product{
		ID:          productID,
		Title:       f.ProductName(),
		Vendor:      f.Company(),
		ProductType: f.ProductCategory(),
		Tags:        tags,
		Variants:    variants,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewInventoryLevel picks a random known variant and emits an inventory
// update for it across a random location. ok is false if no variant has
// been registered yet (e.g. seeding hasn't happened).
func (g *Generator) NewInventoryLevel(locations int) (level model.InventoryLevel, ok bool) {
	ref, ok := g.Registry.RandomVariant(g.Faker)
	if !ok {
		return model.InventoryLevel{}, false
	}

	locationID := int64(g.Faker.IntRange(1, locations))
	available := int32(g.Faker.IntRange(0, 500))

	return model.InventoryLevel{
		InventoryItemID: ref.InventoryItemID,
		SKU:             ref.SKU,
		ProductID:       ref.ProductID,
		LocationID:      locationID,
		Available:       available,
		UpdatedAt:       time.Now().UnixMilli(),
	}, true
}
