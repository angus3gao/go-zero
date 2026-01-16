-- KEYS[1] = dirty:set
-- KEYS[2] = dirty:queue
-- ARGV[1] = key

redis.call('set', ARGV[1], ARGV[2])
redis.call("EXPIRE", ARGV[1], ARGV[3])
if redis.call("SADD", KEYS[1], ARGV[1]) == 1 then
	redis.call("EXPIRE", KEYS[1], 86400)
    redis.call("LPUSH", KEYS[2], ARGV[1])
end
return 1