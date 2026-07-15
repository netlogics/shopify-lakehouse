package gen

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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
	if len(p.Variants) < 1 || len(p.Variants) > 3 {
		t.Errorf("len(Variants) = %d, want 1-3", len(p.Variants))
	}
	if len(p.Tags) == 0 {
		t.Errorf("Tags is empty")
	}
	if p.CreatedAt == 0 || p.UpdatedAt == 0 {
		t.Errorf("CreatedAt/UpdatedAt not set")
	}

	seen := map[int64]bool{}
	for _, v := range p.Variants {
		if v.ID == 0 || v.SKU == "" || v.Price == "" || v.InventoryItemID == 0 {
			t.Errorf("invalid variant: %+v", v)
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

func TestNewInventoryLevelReferencesKnownVariant(t *testing.T) {
	f := gofakeit.New(7)
	reg := NewRegistry()
	g := NewGenerator(f, reg)

	p := g.NewProduct()
	registered := map[int64]variantInfo{}
	for _, v := range p.Variants {
		registered[v.InventoryItemID] = variantInfo{sku: v.SKU, productID: p.ID}
	}

	for i := 0; i < 50; i++ {
		lvl, ok := g.NewInventoryLevel(3)
		if !ok {
			t.Fatal("NewInventoryLevel: expected ok=true with seeded registry")
		}
		info, isKnown := registered[lvl.InventoryItemID]
		if !isKnown {
			t.Fatalf("inventory level references unknown inventory_item_id %d", lvl.InventoryItemID)
		}
		if lvl.SKU != info.sku || lvl.ProductID != info.productID {
			t.Fatalf("inventory level %+v does not match registered variant %+v", lvl, info)
		}
		if lvl.LocationID < 1 || lvl.LocationID > 3 {
			t.Fatalf("LocationID = %d, want in [1,3]", lvl.LocationID)
		}
		if lvl.Available < 0 {
			t.Fatalf("Available = %d, want >= 0", lvl.Available)
		}
	}
}

type variantInfo struct {
	sku       string
	productID int64
}
