-- All validation precedes mutations: Redis Lua runtime failures do not roll back.
local oldtype=redis.call('TYPE',KEYS[1]).ok
local newtype=redis.call('TYPE',KEYS[2]).ok
local familytype=redis.call('TYPE',KEYS[3]).ok
if oldtype~='hash' or familytype~='string' then return 0 end
if newtype~='none' then return -2 end
if redis.call('GET',KEYS[3])~='active' then return 0 end
local ttl=redis.call('PTTL',KEYS[1])
local fttl=redis.call('PTTL',KEYS[3])
if ttl<=0 or fttl<=0 then return 0 end
if redis.call('HGET',KEYS[1],'record')~=ARGV[1] then return 0 end
local used=redis.call('HGET',KEYS[1],'used')
if used~='0' and used~='1' then return 0 end
if used=='1' then
 redis.call('SET',KEYS[3],'revoked','PX',fttl)
 return -1
end
redis.call('HSET',KEYS[2],'record',ARGV[2],'used','0')
redis.call('PEXPIRE',KEYS[2],math.min(ttl,fttl))
redis.call('HSET',KEYS[1],'used','1')
return 1
