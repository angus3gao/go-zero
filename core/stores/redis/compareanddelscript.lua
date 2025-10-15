local val = redis.call('get',KEYS[1])
if val == nil then
	return 1
end

if ARGV[1] == val then
	redis.call('del', KEYS[1])
	return 1
end

return 0