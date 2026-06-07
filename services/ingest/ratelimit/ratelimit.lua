--KEYS[1] = THE RATE LIMIT KEY
--ARGV[1] = LIMIT
--ARGV[2] = WINDOW SECONDS
local current = redis.call("INCR", KEYS[1])
if current == 1 then
	redis.call("EXPIRE", KEYS[1], tonumber(ARGV[2]))
end
return { current, tonumber(ARGV[1]) }
