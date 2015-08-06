package auth

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/gopher-lua"
)

func TestLuaNonExistantFile(t *testing.T) {
	m, err := NewLuaModule("/dev/nonexistingdevice")
	assert.Nil(t, m, "Did not expect a Lua module.")
	assert.NotNil(t, err, "Expected a 'file not found' error.")
}

func TestLuaInvalidScript(t *testing.T) {
	m := NewLuaModuleScript("this is not a lua script")
	Install(m)

	flag, err := Connect(validConnectionRequest())
	assert.False(t, flag, "Expected the connection request to be denied.")
	assert.NotNil(t, err, "Expected an error.")

	luaErr, _ := err.(*lua.ApiError)
	assert.NotNil(t, luaErr, "Expected a Lua syntax error.")
}

func TestLuaScriptWithoutConnectFunction(t *testing.T) {
	m := NewLuaModuleScript("x = 3 + 2")
	Install(m)

	flag, err := Connect(validConnectionRequest())
	assert.False(t, flag, "Expected the connection request to be denied.")
	require.NotNil(t, err, "Expected an error.")
	assert.True(t, strings.Contains(err.Error(), "No function"),
		"Expected a 'no connect function' error.")
}

func TestLuaScriptWithInvalidConnectFUnction(t *testing.T) {
	m := NewLuaModuleScript(`
		function connect(request)
			x = "s" + 5
		end
	`)
	Install(m)

	flag, err := Connect(validConnectionRequest())
	assert.False(t, flag, "Expected the connection request to be denied.")
	assert.NotNil(t, err, "Expected an error.")

	luaErr, _ := err.(*lua.ApiError)
	assert.NotNil(t, luaErr, "Expected a Lua syntax error.")
}

func TestLuaScriptWithConnectWrongReturnType(t *testing.T) {
	m := NewLuaModuleScript(`
	function connect(request)
		return "Not a number!"
	end
	`)
	Install(m)

	flag, err := Connect(validConnectionRequest())
	assert.False(t, flag, "Expected the connection request to be denied.")
	assert.NotNil(t, err, "Expected an error.")
	assert.True(t, strings.Contains(err.Error(), "return a boolean"),
		"Expected a 'expected a boolean' error.")
}

func TestLuaValidScript(t *testing.T) {
	m := NewLuaModuleScript(`
		local http = require("http")
		local json = require("json")

		function connect(request)
			if request:ip() ~= "127.0.0.1" then
				error("IP address does not match.")
			elseif request:gatename() ~= "test_gatename" then
				error("Gatename does not match.")
			elseif request:port() ~= 22 then
				error("Port does not match.")
			elseif request:key() ~= "test_key" then
				error("Key does not match.")
			end
			return true
		end
	`)
	Install(m)

	flag, err := Connect(ConnectionRequest{
		IP:       net.ParseIP("127.0.0.1"),
		Gatename: "test_gatename",
		Port:     22,
		Key:      "test_key",
	})

	assert.True(t, flag, "Expected the connection request to be granted.")
	assert.Nil(t, err, "Did not expect error: %v", err)
}
