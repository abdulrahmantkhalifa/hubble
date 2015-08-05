local http = require("http")

local function url_encode(str)
    if str then
        str = string.gsub(str, "\n", "\r\n")
        str = string.gsub(str, "([^%w %-%_%.%~])",
            function (c) return string.format("%%%02X", string.byte(c)) end)
        str = string.gsub(str, " ", "+")
    end
    return str
end

-- The connect function must be global as it will be called by
-- the Go Lua authorization plugin
function connect(request)
    url = "http://127.0.0.1:3290/authorize"

    url = url .. "?ip=" .. url_encode(request:ip())
    url = url .. "&port=" .. url_encode(tostring(request:port()))
    url = url .. "&gatename=" .. url_encode(request:gatename())
    url = url .. "&key=" .. url_encode(request:key())

    response, err = http.get(url)
    if err then
        error(err)
    end

    if response.status_code == 200 then
        return true
    elseif response.status_code == 404 then
        return false
    else
        error("Unexpected status_code: " .. response.status_code)
    end
end
