local familyKey = KEYS[1]

local familyData = redis.call("GET", familyKey)

if not familyData then
    return 0
end

local family = cjson.decode(familyData)

family.status = "revoked"

redis.call(
    "SET",
    familyKey,
    cjson.encode(family),
    "KEEPTTL"
)

return 1