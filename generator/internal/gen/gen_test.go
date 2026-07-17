package gen

import (
	"strings"
	"testing"
	"time"

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
