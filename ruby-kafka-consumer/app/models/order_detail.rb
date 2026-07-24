require 'active_record'

class OrderDetail < ActiveRecord::Base
  self.table_name = 'order_details'

  validates :topic, :partition, :offset, :consumed_at,
            :order_id, :line_item_id, :title, :name,
            :quantity, :fulfillable_quantity, :current_quantity,
            :price, :total_discount, :fulfillment_service,
            :grams,
            presence: true

  validates :requires_shipping, :taxable, :gift_card, :product_exists,
            inclusion: { in: [true, false] }

  validates :offset, uniqueness: { scope: [:topic, :partition] }
end
