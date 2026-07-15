// Package model defines Go structs matching the Avro schemas in schemas/.
package model

// Variant matches the "Variant" record nested in schemas/product.avsc.
type Variant struct {
	ID              int64  `json:"id" avro:"id"`
	SKU             string `json:"sku" avro:"sku"`
	Price           string `json:"price" avro:"price"`
	InventoryItemID int64  `json:"inventory_item_id" avro:"inventory_item_id"`
}

// Product matches schemas/product.avsc.
type Product struct {
	ID          int64     `json:"id" avro:"id"`
	Title       string    `json:"title" avro:"title"`
	Vendor      string    `json:"vendor" avro:"vendor"`
	ProductType string    `json:"product_type" avro:"product_type"`
	Tags        []string  `json:"tags" avro:"tags"`
	Variants    []Variant `json:"variants" avro:"variants"`
	CreatedAt   int64     `json:"created_at" avro:"created_at"`
	UpdatedAt   int64     `json:"updated_at" avro:"updated_at"`
}

// InventoryLevel matches schemas/inventory_level.avsc.
type InventoryLevel struct {
	InventoryItemID int64  `json:"inventory_item_id" avro:"inventory_item_id"`
	SKU             string `json:"sku" avro:"sku"`
	ProductID       int64  `json:"product_id" avro:"product_id"`
	LocationID      int64  `json:"location_id" avro:"location_id"`
	Available       int32  `json:"available" avro:"available"`
	UpdatedAt       int64  `json:"updated_at" avro:"updated_at"`
}
