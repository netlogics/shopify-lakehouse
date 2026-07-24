require 'active_record'
require 'sqlite3'

ActiveRecord::Base.establish_connection(
  adapter: 'sqlite3',
  database: File.expand_path('../../db/development.sqlite3', __FILE__)
)

# Schema matches schemas/order_detail.avsc and the Shopify REST Admin Order
# API line_items array member.  Kafka envelope columns are prefixed to avoid
# collision with the Shopify field `id`.
ActiveRecord::Schema.define do
  unless table_exists?(:order_details)
    create_table :order_details do |t|
      # Kafka envelope
      t.string   :topic,        null: false
      t.integer  :partition,    null: false
      t.integer  :offset,       null: false
      t.string   :key
      t.datetime :consumed_at,  null: false

      # Shopify order detail (line item) fields — names match the Avro schema
      t.integer  :order_id,                     null: false
      t.integer  :line_item_id,                 null: false  # Avro field: id
      t.integer  :variant_id
      t.integer  :product_id
      t.string   :title,                        null: false
      t.string   :variant_title
      t.string   :name,                         null: false
      t.string   :sku
      t.string   :vendor
      t.integer  :quantity,                     null: false
      t.integer  :fulfillable_quantity,         null: false
      t.integer  :current_quantity,             null: false
      t.string   :price,                        null: false   # decimal string e.g. "74.99"
      t.string   :total_discount,               null: false   # decimal string e.g. "0.00"
      t.string   :fulfillment_service,          null: false
      t.string   :fulfillment_status
      t.integer  :grams,                        null: false
      t.boolean  :requires_shipping,            null: false
      t.boolean  :taxable,                      null: false
      t.boolean  :gift_card,                    null: false
      t.boolean  :product_exists,               null: false
      t.string   :variant_inventory_management
      t.datetime :shopify_created_at
      t.datetime :shopify_updated_at
    end

    add_index :order_details, [:topic, :partition, :offset], unique: true
    add_index :order_details, :order_id
    add_index :order_details, :fulfillment_status
    add_index :order_details, :shopify_created_at
    add_index :order_details, :consumed_at
  end
end
