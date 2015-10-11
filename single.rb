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
s = Stats.new
message = {mark: 'is cool!', number: rand(15)}
s.track({'key' => 'mark', 'value' => message})
