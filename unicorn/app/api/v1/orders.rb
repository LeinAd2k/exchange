# frozen_string_literal: true

module V1
  class Orders < Grape::API
    before do
      authenticate!
    end

    desc 'Get user open orders'
    params do
    end
    get '/open-orders' do
      present []
    end

    desc 'Create a new order'
    params do
    end
    post '/orders' do
    end

    desc 'Cancel order(s)'
    params do
    end
    delete '/orders' do
    end
  end
end
