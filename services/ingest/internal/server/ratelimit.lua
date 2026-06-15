-- KEYS[1] = the rate-limit key, e.g. "rl:user_42:1717750000"
-- ARGV[1] = limit (unused here, kept for symmetry with other scripts)
-- ARGV[2] = window seconds
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], tonumber(ARGV[2]))
end
return current
