local val = redis.call('get',KEYS[1])
if val == nil then
	return 0
end

if ARGV[1] == val then
	redis.call('set', KEYS[1], ARGV[2])
	return 1
end

return 0