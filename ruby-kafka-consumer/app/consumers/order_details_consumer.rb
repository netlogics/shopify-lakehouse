class OrderDetailsConsumer < ApplicationConsumer
  def consume
    messages.each do |message|
      OrderDetail.create!(
        topic: message.topic,
        partition: message.partition,
        offset: message.offset,
        key: message.key,
        payload: message.payload,
        created_at: Time.now.utc
      )
    end
  end
end
