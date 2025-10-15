local score = ARGV[1]
local member = ARGV[2]
local old = redis.call('zscore',KEYS[1],member)

if not old or tonumber(old) < tonumber(score) then
	return redis.call('zadd',KEYS[1], score, member)
end

return 0