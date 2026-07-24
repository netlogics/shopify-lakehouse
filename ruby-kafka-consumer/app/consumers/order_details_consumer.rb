require 'net/http'
require 'json'
require 'avro'
require 'stringio'

# OrderDetailsConsumer decodes Confluent Avro wire-format messages from the
# shopify.order_details topic and persists them as OrderDetail records.
#
# Confluent wire format: 0x00 (magic byte) + 4-byte big-endian schema ID + Avro binary.
# Schemas are fetched from the Schema Registry by ID and cached in memory.
class OrderDetailsConsumer < ApplicationConsumer
  SCHEMA_REGISTRY_URL = ENV.fetch('SCHEMA_REGISTRY_URL', 'http://127.0.0.1:8081')

  def consume
    messages.each do |message|
      record = decode(message.value)

      OrderDetail.create!(
        # Kafka envelope
        topic:        message.topic,
        partition:    message.partition,
        offset:       message.offset,
        key:          message.key,
        consumed_at:  Time.now.utc,

        # Shopify order detail fields; Avro field `id` maps to `line_item_id`
        # to avoid collision with the ActiveRecord primary key column.
        order_id:                     record['order_id'],
        line_item_id:                 record['id'],
        variant_id:                   record['variant_id'],
        product_id:                   record['product_id'],
        title:                        record['title'],
        variant_title:                record['variant_title'],
        name:                         record['name'],
        sku:                          record['sku'],
        vendor:                       record['vendor'],
        quantity:                     record['quantity'],
        fulfillable_quantity:         record['fulfillable_quantity'],
        current_quantity:             record['current_quantity'],
        price:                        record['price'],
        total_discount:               record['total_discount'],
        fulfillment_service:          record['fulfillment_service'],
        fulfillment_status:           record['fulfillment_status'],
        grams:                        record['grams'],
        requires_shipping:            record['requires_shipping'],
        taxable:                      record['taxable'],
        gift_card:                    record['gift_card'],
        product_exists:               record['product_exists'],
        variant_inventory_management: record['variant_inventory_management'],
        shopify_created_at:           parse_timestamp(record['created_at']),
        shopify_updated_at:           parse_timestamp(record['updated_at'])
      )
    end
  end

  private

  # Decode a Confluent Avro wire-format payload.
  # Returns the decoded record as a plain Ruby Hash.
  def decode(payload)
    bytes = payload.b  # ensure binary encoding
    raise "Invalid Confluent magic byte" unless bytes.getbyte(0) == 0

    schema_id = bytes[1, 4].unpack1('N')
    avro_bytes = bytes[5..]

    schema = fetch_schema(schema_id)
    reader = Avro::IO::DatumReader.new(schema)
    decoder = Avro::IO::BinaryDecoder.new(StringIO.new(avro_bytes))
    reader.read(decoder)
  end

  # Fetch and cache an Avro schema from the Schema Registry by numeric ID.
  def fetch_schema(schema_id)
    @schema_cache ||= {}
    @schema_cache[schema_id] ||= begin
      uri = URI("#{SCHEMA_REGISTRY_URL}/schemas/ids/#{schema_id}")
      response = Net::HTTP.get_response(uri)
      raise "Schema Registry error #{response.code} for ID #{schema_id}" unless response.is_a?(Net::HTTPSuccess)

      schema_json = JSON.parse(response.body).fetch('schema')
      Avro::Schema.parse(schema_json)
    end
  end

  def parse_timestamp(value)
    Time.iso8601(value) if value
  end
end
