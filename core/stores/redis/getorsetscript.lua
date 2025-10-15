local old = redis.call('get',KEYS[1])
if old and old ~= "" then
	return old
end

redis.call('set', KEYS[1], ARGV[1])

return ""