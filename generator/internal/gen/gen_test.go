package gen

import (
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"

	"generator/internal/model"
)

func TestNewProduct(t *testing.T) {
	f := gofakeit.New(42)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()

	if p.ID == 0 {
		t.Errorf("Product.ID = 0, want nonzero")
	}
	if p.Title == "" {
		t.Errorf("Product.Title is empty")
	}
	if p.Handle == "" {
		t.Errorf("Product.Handle is empty")
	}
	if p.Status == "" {
		t.Errorf("Product.Status is empty")
	}
	if p.Tags == "" {
		t.Errorf("Product.Tags is empty")
	}
	if p.CreatedAt == "" {
		t.Errorf("Product.CreatedAt is empty")
	}
	if _, err := time.Parse(time.RFC3339, p.CreatedAt); err != nil {
		t.Errorf("Product.CreatedAt not RFC3339: %v", err)
	}
	if len(p.Variants) < 1 || len(p.Variants) > 3 {
		t.Errorf("len(Variants) = %d, want 1-3", len(p.Variants))
	}

	seen := map[int64]bool{}
	for _, v := range p.Variants {
		if v.ID == 0 || v.SKU == "" || v.Price == "" || v.InventoryItemID == 0 {
			t.Errorf("invalid variant: %+v", v)
		}
		if v.ProductID != p.ID {
			t.Errorf("variant.ProductID = %d, want %d", v.ProductID, p.ID)
		}
		if v.Position < 1 {
			t.Errorf("variant.Position = %d, want >= 1", v.Position)
		}
		if !strings.HasPrefix(v.SKU, "SKU-") {
			t.Errorf("variant.SKU = %q, want SKU- prefix", v.SKU)
		}
		if seen[v.ID] {
			t.Errorf("duplicate variant ID %d", v.ID)
		}
		seen[v.ID] = true
	}

	if reg.Len() != len(p.Variants) {
		t.Errorf("registry has %d variants, want %d", reg.Len(), len(p.Variants))
	}
}

func TestNewProductUniqueIDsAcrossCalls(t *testing.T) {
	f := gofakeit.New(1)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	seenProducts := map[int64]bool{}
	seenVariants := map[int64]bool{}
	seenInvItems := map[int64]bool{}

	for i := 0; i < 20; i++ {
		p := g.NewProduct()
		if seenProducts[p.ID] {
			t.Fatalf("duplicate product ID %d", p.ID)
		}
		seenProducts[p.ID] = true
		for _, v := range p.Variants {
			if seenVariants[v.ID] {
				t.Fatalf("duplicate variant ID %d", v.ID)
			}
			seenVariants[v.ID] = true
			if seenInvItems[v.InventoryItemID] {
				t.Fatalf("duplicate inventory item ID %d", v.InventoryItemID)
			}
			seenInvItems[v.InventoryItemID] = true
		}
	}
}

func TestNewInventoryLevelEmptyRegistry(t *testing.T) {
	f := gofakeit.New(7)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	_, ok := g.NewInventoryLevel(3)
	if ok {
		t.Fatal("NewInventoryLevel: expected ok=false on empty registry")
	}
}

func TestNewOrderDetailEmptyCustomerRegistry(t *testing.T) {
	f := gofakeit.New(11)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	// Seed a variant so the only missing prerequisite is a customer.
	p := g.NewProduct()
	variantMap := map[int64]*model.Variant{}
	for i := range p.Variants {
		v := &p.Variants[i]
		variantMap[v.InventoryItemID] = v
	}

	_, ok := g.NewOrderDetail(variantMap, FraudParams{})
	if ok {
		t.Fatal("NewOrderDetail: expected ok=false with no customer registered")
	}
}

func TestNewOrderDetailReferencesKnownCustomer(t *testing.T) {
	f := gofakeit.New(11)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	variantMap := map[int64]*model.Variant{}
	for i := range p.Variants {
		v := &p.Variants[i]
		variantMap[v.InventoryItemID] = v
	}

	knownCustomerIDs := map[int64]bool{}
	for i := 0; i < 5; i++ {
		c := g.NewCustomer()
		knownCustomerIDs[c.ID] = true
	}

	for i := 0; i < 50; i++ {
		detail, ok := g.NewOrderDetail(variantMap, FraudParams{})
		if !ok {
			t.Fatal("NewOrderDetail: expected ok=true with seeded variant and customer registries")
		}
		if detail.CustomerID == nil {
			t.Fatal("OrderDetail.CustomerID is nil, want a known customer ID")
		}
		if !knownCustomerIDs[*detail.CustomerID] {
			t.Fatalf("order detail references unknown customer_id %d", *detail.CustomerID)
		}
	}
}

func TestNewOrderDetailUniqueLineItemIDsAcrossCalls(t *testing.T) {
	f := gofakeit.New(12)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	variantMap := map[int64]*model.Variant{}
	for i := range p.Variants {
		v := &p.Variants[i]
		variantMap[v.InventoryItemID] = v
	}
	g.NewCustomer()

	seen := map[int64]bool{}
	for i := 0; i < 500; i++ {
		detail, ok := g.NewOrderDetail(variantMap, FraudParams{})
		if !ok {
			t.Fatal("NewOrderDetail: expected ok=true with seeded variant and customer registries")
		}
		if detail.ID == 0 {
			t.Fatal("OrderDetail.ID = 0, want nonzero")
		}
		if seen[detail.ID] {
			t.Fatalf("duplicate line_item_id %d", detail.ID)
		}
		seen[detail.ID] = true
	}
}

func TestFraudEpisodeConcentratesOrdersOnTargetCustomer(t *testing.T) {
	f := gofakeit.New(3)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	variantMap := map[int64]*model.Variant{}
	for i := range p.Variants {
		v := &p.Variants[i]
		variantMap[v.InventoryItemID] = v
	}

	other := g.NewCustomer()
	target := g.NewCustomer()
	reg.TriggerFraud(target.ID, time.Minute)

	fraud := FraudParams{TargetWeight: 1} // always target the active episode
	sawFraud := false
	for i := 0; i < 20; i++ {
		detail, ok := g.NewOrderDetail(variantMap, fraud)
		if !ok {
			t.Fatal("NewOrderDetail: expected ok=true")
		}
		if detail.CustomerID == nil || *detail.CustomerID != target.ID {
			t.Fatalf("order attributed to customer %v, want the fraud target %d (other known customer: %d)", detail.CustomerID, target.ID, other.ID)
		}
		if !detail.IsSyntheticFraud {
			t.Fatal("IsSyntheticFraud = false, want true for an order attributed to an active fraud episode")
		}
		if detail.FraudPattern == nil || *detail.FraudPattern != fraudPatternVelocityBurst {
			t.Fatalf("FraudPattern = %v, want %q", detail.FraudPattern, fraudPatternVelocityBurst)
		}
		if detail.Quantity < 50 || detail.Quantity > 200 {
			t.Fatalf("Quantity = %d, want anomalous range [50,200]", detail.Quantity)
		}
		sawFraud = true
	}
	if !sawFraud {
		t.Fatal("no fraud-labeled orders generated")
	}
}

func TestNewOrderDetailFraudDisabledByDefault(t *testing.T) {
	f := gofakeit.New(5)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	variantMap := map[int64]*model.Variant{}
	for i := range p.Variants {
		v := &p.Variants[i]
		variantMap[v.InventoryItemID] = v
	}
	g.NewCustomer()

	for i := 0; i < 50; i++ {
		detail, ok := g.NewOrderDetail(variantMap, FraudParams{})
		if !ok {
			t.Fatal("NewOrderDetail: expected ok=true")
		}
		if detail.IsSyntheticFraud {
			t.Fatal("IsSyntheticFraud = true with zero-value FraudParams, want fraud injection disabled")
		}
		if detail.FraudPattern != nil {
			t.Fatalf("FraudPattern = %v, want nil with fraud injection disabled", *detail.FraudPattern)
		}
	}
	if reg.HasActiveFraud() {
		t.Fatal("HasActiveFraud() = true with zero-value FraudParams, want no episode ever triggered")
	}
}

func TestMaybeTriggerFraudDoesNotOverlapEpisodes(t *testing.T) {
	f := gofakeit.New(9)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	for i := 0; i < 10; i++ {
		g.NewCustomer()
	}
	reg.TriggerFraud(1, time.Minute)

	fraud := FraudParams{InjectionProbability: 1, EpisodeDuration: time.Minute}
	for i := 0; i < 20; i++ {
		g.maybeTriggerFraud(fraud)
	}

	active := 0
	reg.mu.Lock()
	for _, until := range reg.fraudUntil {
		if time.Now().Before(until) {
			active++
		}
	}
	reg.mu.Unlock()
	if active != 1 {
		t.Fatalf("active fraud episodes = %d, want exactly 1 (no overlapping episodes)", active)
	}
}

func TestNewInventoryLevelReferencesKnownVariant(t *testing.T) {
	f := gofakeit.New(7)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	knownItemIDs := map[int64]bool{}
	for _, v := range p.Variants {
		knownItemIDs[v.InventoryItemID] = true
	}

	for i := 0; i < 50; i++ {
		lvl, ok := g.NewInventoryLevel(3)
		if !ok {
			t.Fatal("NewInventoryLevel: expected ok=true with seeded registry")
		}
		if !knownItemIDs[lvl.InventoryItemID] {
			t.Fatalf("inventory level references unknown inventory_item_id %d", lvl.InventoryItemID)
		}
		if lvl.LocationID < 1 || lvl.LocationID > 3 {
			t.Fatalf("LocationID = %d, want in [1,3]", lvl.LocationID)
		}
		if lvl.Available == nil {
			t.Fatalf("Available is nil")
		}
		if *lvl.Available < 0 {
			t.Fatalf("Available = %d, want >= 0", *lvl.Available)
		}
		if _, err := time.Parse(time.RFC3339, lvl.UpdatedAt); err != nil {
			t.Errorf("InventoryLevel.UpdatedAt not RFC3339: %v", err)
		}
	}
}
