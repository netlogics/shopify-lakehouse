# frozen_string_literal: true

ENV['RAILS_ENV'] ||= 'development'

require 'karafka'
require_relative 'app/models/database'
require_relative 'app/models/order_detail'
require_relative 'app/consumers/application_consumer'
require_relative 'app/consumers/order_details_consumer'

class KarafkaApp < Karafka::App
  setup do |config|
    config.kafka = { 'bootstrap.servers': '127.0.0.1:9092' }
    config.client_id = 'order_details_consumer'
  end

  routes.draw do
    topic :'order-details' do
      consumer OrderDetailsConsumer
    end
  end
end

KarafkaApp.boot
