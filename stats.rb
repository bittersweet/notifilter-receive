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
    # if @backlog.size >= @batch_size
      flush
    # end
  end

  def flush
    @backlog.each do |item|
      message = item.to_json
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
    100.times do |i|
      s.track({'key' => 'mark', 'value' => rand(100).to_s})
    end
  end
end

jobs.map(&:join)
