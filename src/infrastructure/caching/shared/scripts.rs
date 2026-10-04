// Preflight all types and values before any mutation: Redis scripts do not roll back.
const REVISION_PREFLIGHT: &str = r#"
for i = 1, #KEYS do
    local kind = redis.call('TYPE', KEYS[i]).ok
    if kind ~= 'none' and kind ~= 'string' then
        return redis.error_reply('invalid revision type')
    end
    if kind == 'string' and redis.call('GET', KEYS[i]) == '' then
        return redis.error_reply('invalid revision value')
    end
    if not ARGV[i] or ARGV[i] == '' then
        return redis.error_reply('invalid revision argument')
    end
end
"#;

pub(crate) fn initialize() -> redis::Script {
    redis::Script::new(&format!(
        "{}{}",
        REVISION_PREFLIGHT,
        r#"
local values = {}
for i = 1, #KEYS do
    local value = redis.call('GET', KEYS[i])
    if not value then
        value = ARGV[i]
        redis.call('SET', KEYS[i], value)
    end
    values[i] = value
end
return values
"#
    ))
}

pub(crate) fn invalidate() -> redis::Script {
    redis::Script::new(&format!(
        "{}{}",
        REVISION_PREFLIGHT,
        r#"
for i = 1, #KEYS do
    redis.call('SET', KEYS[i], ARGV[i])
end
return 1
"#
    ))
}

// The final key is the entry; preceding keys are revision dependencies.
pub(crate) fn read() -> redis::Script {
    redis::Script::new(
        r#"
for i = 1, #KEYS - 1 do
    local kind = redis.call('TYPE', KEYS[i]).ok
    if kind ~= 'none' and kind ~= 'string' then
        return redis.error_reply('invalid revision type')
    end
    local value = redis.call('GET', KEYS[i])
    if value == '' then return redis.error_reply('invalid revision value') end
end
for i = 1, #KEYS - 1 do
    if redis.call('GET', KEYS[i]) ~= ARGV[i] then return false end
end
return redis.call('GET', KEYS[#KEYS])
"#,
    )
}

pub(crate) fn store() -> redis::Script {
    redis::Script::new(
        r#"
for i = 1, #KEYS - 1 do
    local kind = redis.call('TYPE', KEYS[i]).ok
    if kind ~= 'none' and kind ~= 'string' then
        return redis.error_reply('invalid revision type')
    end
    local value = redis.call('GET', KEYS[i])
    if value == '' then return redis.error_reply('invalid revision value') end
end
for i = 1, #KEYS - 1 do
    if redis.call('GET', KEYS[i]) ~= ARGV[i] then return 0 end
end
local kind = redis.call('TYPE', KEYS[#KEYS]).ok
if kind ~= 'none' and kind ~= 'string' then
    return redis.error_reply('invalid entry type')
end
redis.call('SET', KEYS[#KEYS], ARGV[#KEYS], 'PX', ARGV[#KEYS + 1])
return 1
"#,
    )
}
