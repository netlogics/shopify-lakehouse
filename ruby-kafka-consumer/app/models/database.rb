require 'active_record'
require 'sqlite3'

# Establish connection
ActiveRecord::Base.establish_connection(
  adapter: 'sqlite3',
  database: File.expand_path('../../db/development.sqlite3', __FILE__)
)

# Run migrations inline (only if table doesn't exist)
ActiveRecord::Schema.define do
  unless table_exists?(:order_details)
    create_table :order_details do |t|
      t.string :topic
      t.integer :partition
      t.integer :offset
      t.string :key
      t.text :payload
      t.datetime :created_at
    end

    add_index :order_details, [:topic, :partition, :offset], unique: true
    add_index :order_details, :created_at
  end
end
