
local oldKey = KEYS[1]
local familyKey = KEYS[2]
local newKey = KEYS[3]

local newTokenData = ARGV[1]
local newTTL = tonumber(ARGV[2])

--  Validate requested TTL
if not newTTL or newTTL <= 0 then
    return -5
end

-- Retrieve old refresh token
local oldData = redis.call("GET", oldKey)

if not oldData then
    return -1
end

local oldToken = cjson.decode(oldData)

-- Retrieve refresh family
local familyData = redis.call("GET", familyKey)

if not familyData then
    return -2
end

local family = cjson.decode(familyData)

-- Detect reuse before checking other token properties
if oldToken.status == "used" then
    family.status = "revoked"

    redis.call(
        "SET",
        familyKey,
        cjson.encode(family),
        "KEEPTTL"
    )

    return -3
end

if family.status ~= "active" then
    return -2
end

if oldToken.status ~= "active" then
    return -1
end

-- Verify token belongs to this family and user
if oldToken.family_id ~= family.family_id
    or oldToken.user_id ~= family.user_id then
    return -1
end

-- Validate the new token record
local newToken = cjson.decode(newTokenData)

if newToken.family_id ~= family.family_id
    or newToken.user_id ~= family.user_id
    or newToken.status ~= "active" then
    return -1
end

-- Prevent accidental overwriting of an existing token
if redis.call("EXISTS", newKey) == 1 then
    return -4
end

-- Preserve the original session expiry
local familyTTL = redis.call("PTTL", familyKey)

if familyTTL <= 0 then
    return -2
end

if newTTL > familyTTL then
    newTTL = familyTTL
end

-- Mark old token as used
oldToken.status = "used"

redis.call(
    "SET",
    oldKey,
    cjson.encode(oldToken),
    "KEEPTTL"
)

-- Store new refresh token
redis.call(
    "SET",
    newKey,
    newTokenData,
    "PX",
    newTTL
)

return 1
