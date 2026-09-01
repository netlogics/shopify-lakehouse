// Package gen produces fake Shopify product and inventory events matching the
// shape of the Shopify REST Admin API / webhook payloads.
package gen

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"generator/internal/model"
)

// VariantRef tracks inventory_item_id → product_id for referential integrity
// when generating inventory level events.
type VariantRef struct {
	InventoryItemID int64
	ProductID       int64
}

// Registry tracks every variant and customer seen so far.
type Registry struct {
	mu                  sync.Mutex
	variants            []VariantRef
	customers           []int64
	nextProductID       int64
	nextVariantID       int64
	nextInventoryItemID int64
	nextCustomerID      int64
	usedHandles         map[string]struct{}
	fraudUntil          map[int64]time.Time
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{
		usedHandles: make(map[string]struct{}),
		fraudUntil:  make(map[int64]time.Time),
	}
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

// nextCustomer allocates the next customer ID and registers it for later
// random lookup by RandomCustomer, atomically under the registry lock.
func (r *Registry) nextCustomer() int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextCustomerID++
	id := r.nextCustomerID
	r.customers = append(r.customers, id)
	return id
}

// RandomCustomer picks a uniformly random known customer ID. ok is false if
// no customer has been registered yet.
func (r *Registry) RandomCustomer(f *gofakeit.Faker) (id int64, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.customers) == 0 {
		return 0, false
	}
	idx := f.IntRange(0, len(r.customers)-1)
	return r.customers[idx], true
}

// TriggerFraud marks customerID as being in an active synthetic-fraud
// episode until duration from now. Concentrating subsequent order-detail
// generation onto this one customer (see RandomFraudulentCustomer) produces
// a detectable velocity burst rather than an isolated, unlabeled anomaly.
func (r *Registry) TriggerFraud(customerID int64, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fraudUntil[customerID] = time.Now().Add(duration)
}

// HasActiveFraud reports whether any customer is currently within an active
// synthetic-fraud episode. Used to avoid starting overlapping episodes,
// which would dilute the velocity signal across multiple customers.
func (r *Registry) HasActiveFraud() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for _, until := range r.fraudUntil {
		if now.Before(until) {
			return true
		}
	}
	return false
}

// RandomFraudulentCustomer picks a uniformly random customer currently
// within an active fraud episode. ok is false if none is active.
func (r *Registry) RandomFraudulentCustomer(f *gofakeit.Faker) (id int64, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	var active []int64
	for cid, until := range r.fraudUntil {
		if now.Before(until) {
			active = append(active, cid)
		}
	}
	if len(active) == 0 {
		return 0, false
	}
	idx := f.IntRange(0, len(active)-1)
	return active[idx], true
}

// UniqueHandle ensures the given base handle is unique by appending a
// numeric suffix if needed. Returns "base", "base-1", "base-2", ...
func (r *Registry) UniqueHandle(base string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.usedHandles[base]; !exists {
		r.usedHandles[base] = struct{}{}
		return base
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, exists := r.usedHandles[candidate]; !exists {
			r.usedHandles[candidate] = struct{}{}
			return candidate
		}
	}
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
	customerStates     = []string{"enabled", "enabled", "enabled", "disabled", "pending"}
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
	productHandle := g.Registry.UniqueHandle(handle(title))
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
		EventID:     model.EventID(uuid.New().String()),
		ID:          productID,
		Title:       title,
		BodyHTML:    fmt.Sprintf("<p>%s</p>", f.LoremIpsumSentence(8)),
		Vendor:      f.Company(),
		ProductType: f.ProductCategory(),
		Handle:      productHandle,
		Status:      status,
		Tags:        tagsStr,
		CreatedAt:   shopifyTime(createdAt),
		UpdatedAt:   shopifyTime(updatedAt),
		PublishedAt: publishedAtPtr,
		Variants:    variants,
	}
}

// FraudParams controls synthetic fraud injection for NewOrderDetail. See
// generator/internal/config.FraudConfig for the YAML/env-configurable
// source of these values.
type FraudParams struct {
	// InjectionProbability is the chance, per call, of starting a new fraud
	// episode when none is currently active.
	InjectionProbability float64
	// EpisodeDuration is how long a triggered episode concentrates order
	// volume onto its target customer.
	EpisodeDuration time.Duration
	// TargetWeight is the chance a call is attributed to a customer with an
	// active episode rather than a uniformly random customer.
	TargetWeight float64
}

const fraudPatternVelocityBurst = "velocity_burst"

// maybeTriggerFraud starts a new fraud episode on a random known customer
// with probability fraud.InjectionProbability, but only if no episode is
// currently active (overlapping episodes would dilute the velocity signal
// across multiple customers instead of producing one detectable burst).
func (g *Generator) maybeTriggerFraud(fraud FraudParams) {
	if fraud.InjectionProbability <= 0 || g.Registry.HasActiveFraud() {
		return
	}
	if g.Faker.Float64Range(0, 1) >= fraud.InjectionProbability {
		return
	}
	if customerID, ok := g.Registry.RandomCustomer(g.Faker); ok {
		g.Registry.TriggerFraud(customerID, fraud.EpisodeDuration)
	}
}

// selectOrderCustomer decides which customer an order-detail should be
// attributed to. With probability fraud.TargetWeight it targets a customer
// currently in an active fraud episode, concentrating volume onto them;
// otherwise (or if none is active) it falls back to a uniformly random
// known customer. ok is false if no customer has been registered yet.
func (g *Generator) selectOrderCustomer(fraud FraudParams) (customerID int64, isFraud bool, ok bool) {
	if fraud.TargetWeight > 0 && g.Faker.Float64Range(0, 1) < fraud.TargetWeight {
		if id, active := g.Registry.RandomFraudulentCustomer(g.Faker); active {
			return id, true, true
		}
	}
	id, ok := g.Registry.RandomCustomer(g.Faker)
	return id, false, ok
}

// NewOrderDetail picks a random known variant and a customer (see
// selectOrderCustomer), and emits an order detail (line item) event for a
// fake order attributed to that customer. The variantMap provides O(1)
// lookup by inventory_item_id. ok is false if no variant or no customer has
// been registered yet. fraud controls synthetic fraud injection; pass a
// zero-value FraudParams to disable it entirely.
func (g *Generator) NewOrderDetail(variantMap map[int64]*model.Variant, fraud FraudParams) (detail model.OrderDetail, ok bool) {
	ref, ok := g.Registry.RandomVariant(g.Faker)
	if !ok {
		return model.OrderDetail{}, false
	}

	g.maybeTriggerFraud(fraud)
	customerID, isFraud, ok := g.selectOrderCustomer(fraud)
	if !ok {
		return model.OrderDetail{}, false
	}

	// O(1) lookup by inventory_item_id.
	variant := variantMap[ref.InventoryItemID]

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

	var fraudPattern *string
	if isFraud {
		// Anomalously large quantity (vs. the normal 1-10 range) at the
		// variant's real price, so both a per-customer order-count velocity
		// query and a per-customer SUM(quantity*price) query can catch it.
		quantity = int32(f.IntRange(50, 200))
		pattern := fraudPatternVelocityBurst
		fraudPattern = &pattern
	}

	return model.OrderDetail{
		EventID:                    model.EventID(uuid.New().String()),
		OrderID:                    orderID,
		ID:                         lineItemID,
		VariantID:                  variantID,
		ProductID:                  productID,
		CustomerID:                 &customerID,
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
		IsSyntheticFraud:           isFraud,
		FraudPattern:               fraudPattern,
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
		EventID:         model.EventID(uuid.New().String()),
		InventoryItemID: ref.InventoryItemID,
		LocationID:      locationID,
		Available:       &available,
		UpdatedAt:       shopifyTime(time.Now()),
	}, true
}

// NewCustomer creates a fake customer event. IDs are monotonically increasing
// via a dedicated counter (independent of product/variant IDs), and each new
// ID is registered so NewOrderDetail can later reference a real customer.
func (g *Generator) NewCustomer() model.Customer {
	f := g.Faker
	now := time.Now()
	createdAt := now.Add(-time.Duration(f.IntRange(1, 365*24)) * time.Hour)
	updatedAt := createdAt.Add(time.Duration(f.IntRange(0, 72)) * time.Hour)
	if updatedAt.After(now) {
		updatedAt = now
	}

	// Monotonically increasing customer ID, registered for later random
	// lookup by order-detail generation (see RandomCustomer).
	customerID := g.Registry.nextCustomer()

	firstName := f.FirstName()
	lastName := f.LastName()
	email := strings.ToLower(fmt.Sprintf("%s.%s%d@example.com", firstName, lastName, customerID))
	state := customerStates[f.IntRange(0, len(customerStates)-1)]

	var phone *string
	if f.Bool() {
		p := fmt.Sprintf("+%d-%d-%d-%d", f.IntRange(1, 9), f.IntRange(100, 999), f.IntRange(100, 999), f.IntRange(1000, 9999))
		phone = &p
	}

	var tags *string
	if f.IntRange(0, 3) > 0 {
		numTags := f.IntRange(1, 3)
		tagWords := make([]string, numTags)
		for i := range tagWords {
			tagWords[i] = f.Word()
		}
		t := strings.Join(tagWords, ", ")
		tags = &t
	}

	return model.Customer{
		EventID:       model.EventID(uuid.New().String()),
		ID:            customerID,
		Email:         &email,
		FirstName:     &firstName,
		LastName:      &lastName,
		Phone:         phone,
		State:         state,
		VerifiedEmail: f.Bool(),
		Tags:          tags,
		CreatedAt:     shopifyTime(createdAt),
		UpdatedAt:     shopifyTime(updatedAt),
	}
}
