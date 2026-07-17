// Package model defines Go structs matching the Avro schemas in schemas/.
// Field names and types mirror the Shopify REST Admin API / webhook payloads.
package model

// Variant matches the "Variant" record nested in schemas/product.avsc.
type Variant struct {
	ID                  int64   `json:"id" avro:"id"`
	ProductID           int64   `json:"product_id" avro:"product_id"`
	Title               string  `json:"title" avro:"title"`
	Price               string  `json:"price" avro:"price"`
	SKU                 string  `json:"sku" avro:"sku"`
	Position            int32   `json:"position" avro:"position"`
	InventoryPolicy     string  `json:"inventory_policy" avro:"inventory_policy"`
	CompareAtPrice      *string `json:"compare_at_price" avro:"compare_at_price"`
	FulfillmentService  string  `json:"fulfillment_service" avro:"fulfillment_service"`
	InventoryManagement *string `json:"inventory_management" avro:"inventory_management"`
	Option1             *string `json:"option1" avro:"option1"`
	Option2             *string `json:"option2" avro:"option2"`
	Option3             *string `json:"option3" avro:"option3"`
	Taxable             bool    `json:"taxable" avro:"taxable"`
	Barcode             *string `json:"barcode" avro:"barcode"`
	Grams               int32   `json:"grams" avro:"grams"`
	Weight              float64 `json:"weight" avro:"weight"`
	WeightUnit          string  `json:"weight_unit" avro:"weight_unit"`
	InventoryItemID     int64   `json:"inventory_item_id" avro:"inventory_item_id"`
	InventoryQuantity   int32   `json:"inventory_quantity" avro:"inventory_quantity"`
	RequiresShipping    bool    `json:"requires_shipping" avro:"requires_shipping"`
	CreatedAt           string  `json:"created_at" avro:"created_at"`
	UpdatedAt           string  `json:"updated_at" avro:"updated_at"`
}

// Product matches schemas/product.avsc and the Shopify REST API product object.
type Product struct {
	ID          int64    `json:"id" avro:"id"`
	Title       string   `json:"title" avro:"title"`
	BodyHTML    string   `json:"body_html" avro:"body_html"`
	Vendor      string   `json:"vendor" avro:"vendor"`
	ProductType string   `json:"product_type" avro:"product_type"`
	Handle      string   `json:"handle" avro:"handle"`
	Status      string   `json:"status" avro:"status"`
	Tags        string   `json:"tags" avro:"tags"`
	CreatedAt   string   `json:"created_at" avro:"created_at"`
	UpdatedAt   string   `json:"updated_at" avro:"updated_at"`
	PublishedAt *string  `json:"published_at" avro:"published_at"`
	Variants    []Variant `json:"variants" avro:"variants"`
}

// InventoryLevel matches schemas/inventory_level.avsc and the Shopify REST API
// inventory_level object. Note: sku and product_id are not part of the API
// response — use a join with products/variants to enrich downstream.
type InventoryLevel struct {
	InventoryItemID int64  `json:"inventory_item_id" avro:"inventory_item_id"`
	LocationID      int64  `json:"location_id" avro:"location_id"`
	Available       *int32 `json:"available" avro:"available"`
	UpdatedAt       string `json:"updated_at" avro:"updated_at"`
}
