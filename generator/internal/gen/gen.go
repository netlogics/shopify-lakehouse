// Package gen produces fake Shopify product and inventory events matching the
// shape of the Shopify REST Admin API / webhook payloads.
package gen

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"generator/internal/model"
)

// VariantRef tracks inventory_item_id → product_id for referential integrity
// when generating inventory level events.
type VariantRef struct {
	InventoryItemID int64
	ProductID       int64
}

// Registry tracks every variant seen so far.
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

// Generator produces fake product and inventory events.
type Generator struct {
	Faker    *gofakeit.Faker
	Registry *Registry
}

// NewGenerator builds a Generator backed by the given faker and registry.
func NewGenerator(f *gofakeit.Faker, reg *Registry) *Generator {
	return &Generator{Faker: f, Registry: reg}
}

var (
	inventoryPolicies  = []string{"deny", "continue"}
	weightUnits        = []string{"lb", "kg", "g", "oz"}
	productStatuses    = []string{"active", "active", "active", "draft", "archived"}
	inventoryMgmt      = "shopify"
	fulfillmentService = "manual"
)

func strPtr(s string) *string { return &s }

// handle converts a product title to a URL-safe handle, e.g. "Red T-Shirt" → "red-t-shirt".
func handle(title string) string {
	h := strings.ToLower(title)
	h = strings.ReplaceAll(h, " ", "-")
	// strip non-alphanumeric except hyphens
	var b strings.Builder
	for _, r := range h {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func shopifyTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// NewProduct creates a fake product with 1-3 variants, registering the new
// variants so later inventory events can reference them.
func (g *Generator) NewProduct() model.Product {
	f := g.Faker
	numVariants := f.IntRange(1, 3)
	productID, variantIDs, inventoryItemIDs := g.Registry.nextIDs(numVariants)

	now := time.Now()
	createdAt := now.Add(-time.Duration(f.IntRange(1, 365*24)) * time.Hour)
	updatedAt := createdAt.Add(time.Duration(f.IntRange(0, 72)) * time.Hour)
	if updatedAt.After(now) {
		updatedAt = now
	}
	publishedAt := createdAt.Add(time.Duration(f.IntRange(0, 48)) * time.Hour)

	title := f.ProductName()
	status := productStatuses[f.IntRange(0, len(productStatuses)-1)]

	// Tags: 1-4 comma-separated words, matching Shopify REST format.
	numTags := f.IntRange(1, 4)
	tags := make([]string, numTags)
	for i := range tags {
		tags[i] = strings.Title(f.LoremIpsumWord()) //nolint:staticcheck
	}
	tagsStr := strings.Join(tags, ", ")

	variants := make([]model.Variant, numVariants)
	refs := make([]VariantRef, numVariants)

	// Generate option names: up to 3 per product (Color, Size, Material).
	optionNames := []string{"Color", "Size", "Material"}
	numOptions := f.IntRange(1, 3)
	optionValues := make([]string, numOptions)
	for i := 0; i < numOptions; i++ {
		switch optionNames[i] {
		case "Color":
			optionValues[i] = f.Color()
		case "Size":
			optionValues[i] = []string{"XS", "S", "M", "L", "XL", "XXL"}[f.IntRange(0, 5)]
		case "Material":
			optionValues[i] = []string{"Cotton", "Polyester", "Wool", "Linen", "Silk"}[f.IntRange(0, 4)]
		}
	}

	for i := 0; i < numVariants; i++ {
		sku := fmt.Sprintf("SKU-%d-%d", productID, variantIDs[i])
		price := fmt.Sprintf("%.2f", f.Price(5, 500))
		weight := f.Float64Range(0.1, 10.0)
		weightUnit := weightUnits[f.IntRange(0, len(weightUnits)-1)]
		grams := int32(weight * 453.592) // approximate; exact conversion varies by unit
		invPolicy := inventoryPolicies[f.IntRange(0, len(inventoryPolicies)-1)]
		qty := int32(f.IntRange(0, 500))

		var opt1, opt2, opt3 *string
		if numOptions >= 1 {
			opt1 = strPtr(optionValues[0])
		}
		if numOptions >= 2 {
			opt2 = strPtr(optionValues[1])
		}
		if numOptions >= 3 {
			opt3 = strPtr(optionValues[2])
		}

		varTitle := "Default Title"
		if opt1 != nil {
			varTitle = *opt1
			if opt2 != nil {
				varTitle += " / " + *opt2
			}
		}

		var barcode *string
		if f.Bool() {
			bc := fmt.Sprintf("%012d", 100000000000+int64(f.IntRange(0, 899999999)))
			barcode = &bc
		}

		var compareAtPrice *string
		if f.Bool() {
			cap := fmt.Sprintf("%.2f", f.Price(10, 600))
			compareAtPrice = &cap
		}

		variants[i] = model.Variant{
			ID:                  variantIDs[i],
			ProductID:           productID,
			Title:               varTitle,
			Price:               price,
			SKU:                 sku,
			Position:            int32(i + 1),
			InventoryPolicy:     invPolicy,
			CompareAtPrice:      compareAtPrice,
			FulfillmentService:  fulfillmentService,
			InventoryManagement: strPtr(inventoryMgmt),
			Option1:             opt1,
			Option2:             opt2,
			Option3:             opt3,
			Taxable:             true,
			Barcode:             barcode,
			Grams:               grams,
			Weight:              weight,
			WeightUnit:          weightUnit,
			InventoryItemID:     inventoryItemIDs[i],
			InventoryQuantity:   qty,
			RequiresShipping:    true,
			CreatedAt:           shopifyTime(createdAt),
			UpdatedAt:           shopifyTime(updatedAt),
		}
		refs[i] = VariantRef{
			InventoryItemID: inventoryItemIDs[i],
			ProductID:       productID,
		}
	}
	g.Registry.addVariants(refs)

	var publishedAtPtr *string
	if status == "active" {
		s := shopifyTime(publishedAt)
		publishedAtPtr = &s
	}

	return model.Product{
		ID:          productID,
		Title:       title,
		BodyHTML:    fmt.Sprintf("<p>%s</p>", f.LoremIpsumSentence(8)),
		Vendor:      f.Company(),
		ProductType: f.ProductCategory(),
		Handle:      handle(title),
		Status:      status,
		Tags:        tagsStr,
		CreatedAt:   shopifyTime(createdAt),
		UpdatedAt:   shopifyTime(updatedAt),
		PublishedAt: publishedAtPtr,
		Variants:    variants,
	}
}

// NewOrderDetail picks a random known variant and emits an order detail (line
// item) event for a fake order. ok is false if no variant has been registered
// yet.
func (g *Generator) NewOrderDetail(products []model.Product) (detail model.OrderDetail, ok bool) {
	ref, ok := g.Registry.RandomVariant(g.Faker)
	if !ok {
		return model.OrderDetail{}, false
	}

	// Find the matching variant to carry its fields through.
	var variant *model.Variant
	for i := range products {
		for j := range products[i].Variants {
			if products[i].Variants[j].InventoryItemID == ref.InventoryItemID {
				v := products[i].Variants[j]
				variant = &v
				break
			}
		}
		if variant != nil {
			break
		}
	}

	f := g.Faker
	now := time.Now()
	createdAt := now.Add(-time.Duration(f.IntRange(1, 30*24)) * time.Hour)
	updatedAt := createdAt.Add(time.Duration(f.IntRange(0, 48)) * time.Hour)
	if updatedAt.After(now) {
		updatedAt = now
	}

	quantity := int32(f.IntRange(1, 10))
	orderID := int64(f.IntRange(1_000_000, 9_999_999))
	lineItemID := int64(f.IntRange(1_000_000_000, 9_999_999_999))

	price := fmt.Sprintf("%.2f", f.Price(5, 500))
	grams := int32(f.IntRange(100, 5000))
	title := f.ProductName()
	sku := fmt.Sprintf("SKU-%d-%d", ref.ProductID, ref.InventoryItemID)
	vendor := f.Company()
	variantTitle := "Default Title"
	variantInvMgmt := "shopify"

	var variantID, productID *int64
	if variant != nil {
		price = variant.Price
		grams = variant.Grams
		sku = variant.SKU
		if variant.Option1 != nil {
			variantTitle = *variant.Option1
			if variant.Option2 != nil {
				variantTitle += " / " + *variant.Option2
			}
		}
		variantID = &variant.ID
		productID = &variant.ProductID
		if variant.InventoryManagement != nil {
			variantInvMgmt = *variant.InventoryManagement
		}
	}

	return model.OrderDetail{
		OrderID:                    orderID,
		ID:                         lineItemID,
		VariantID:                  variantID,
		ProductID:                  productID,
		Title:                      title,
		VariantTitle:               &variantTitle,
		Name:                       fmt.Sprintf("%s - %s", title, variantTitle),
		SKU:                        &sku,
		Vendor:                     &vendor,
		Quantity:                   quantity,
		FulfillableQuantity:        quantity,
		CurrentQuantity:            quantity,
		Price:                      price,
		TotalDiscount:              "0.00",
		FulfillmentService:         fulfillmentService,
		FulfillmentStatus:          nil,
		Grams:                      grams,
		RequiresShipping:           true,
		Taxable:                    true,
		GiftCard:                   false,
		ProductExists:              true,
		VariantInventoryManagement: &variantInvMgmt,
		CreatedAt:                  shopifyTime(createdAt),
		UpdatedAt:                  shopifyTime(updatedAt),
	}, true
}

// NewInventoryLevel picks a random known variant and emits an inventory update
// for a random location. ok is false if no variant has been registered yet.
func (g *Generator) NewInventoryLevel(locations int) (level model.InventoryLevel, ok bool) {
	ref, ok := g.Registry.RandomVariant(g.Faker)
	if !ok {
		return model.InventoryLevel{}, false
	}

	locationID := int64(g.Faker.IntRange(1, locations))
	available := int32(g.Faker.IntRange(0, 500))

	return model.InventoryLevel{
		InventoryItemID: ref.InventoryItemID,
		LocationID:      locationID,
		Available:       &available,
		UpdatedAt:       shopifyTime(time.Now()),
	}, true
}
