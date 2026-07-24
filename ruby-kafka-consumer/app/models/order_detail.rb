require 'active_record'

class OrderDetail < ActiveRecord::Base
  self.table_name = 'order_details'

  validates :topic, :partition, :offset, presence: true
  validates :offset, uniqueness: { scope: [:topic, :partition] }
end
