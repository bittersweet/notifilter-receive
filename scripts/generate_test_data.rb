require 'socket'
require 'json'

class Stats
  def initialize
    @batch_size = 20
    @backlog = []
  end

  def socket
    Thread.current[:statsd_socket] ||= UDPSocket.new
  end

  def track(message)
    @backlog << message
    # uncomment to enable buffering messages per 20
    # if @backlog.size >= @batch_size
    #   flush
    # end
    flush
  end

  def flush
    @backlog.each do |item|
      message = item.to_json
      puts message
      socket.send(message, 0, "127.0.0.1", 8000)
    end
    @backlog.clear
  end
end

socket = UDPSocket.new
jobs = []
10.times do
  jobs << Thread.new do
    s = Stats.new
    1000.times do |i|
      data = { user_id: rand(10), created_at: Time.now }
      s.track({'identifier' => 'signup', 'data' => data})
    end
  end
end

jobs.map(&:join)
