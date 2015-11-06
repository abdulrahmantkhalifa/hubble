package auth

import (
	"errors"
	"net/http"
	"os"

	"github.com/cjoudrey/gluahttp"
	"github.com/layeh/gopher-json"
	"github.com/yuin/gopher-lua"
)

type luaAuthModule struct {
	file   string
	script string
}

func NewLuaModule(file string) (AuthorizationModule, error) {
	if _, err := os.Stat(file); err != nil {
		return nil, err
	}

	m := luaAuthModule{
		file: file,
	}

	return m, nil
}

func NewLuaModuleScript(script string) AuthorizationModule {
	return luaAuthModule{
		script: script,
	}
}

func (auth luaAuthModule) Connect(r ConnectionRequest) (bool, error) {
	l := lua.NewState()
	defer l.Close()

	// Give lua script access to Go http and json
	l.PreloadModule("http", gluahttp.NewHttpModule(&http.Client{}).Loader)
	json.Preload(l)
	auth.registerConnectionRequestType(l)

	var err error
	if auth.script != "" {
		err = l.DoString(auth.script)
	} else {
		err = l.DoFile(auth.file)
	}
	if err != nil {
		return false, err
	}

	connect := l.GetGlobal("connect")
	if connect.Type() != lua.LTFunction {
		return false, errors.New("No function 'connect' in Lua module.")
	}

	param := lua.P{
		Fn:      connect,
		NRet:    1,
		Protect: true,
	}

	luaReq := l.NewUserData()
	luaReq.Value = r
	l.SetMetatable(luaReq, l.GetTypeMetatable("ConnectionRequest"))

	if err := l.CallByParam(param, luaReq); err != nil {
		return false, err
	}

	ret := l.Get(-1)
	l.Pop(1)

	if ret.Type() != lua.LTBool {
		return false, errors.New("Function 'connect' does not return a boolean.")
	}

	return ret == lua.LTrue, nil
}

func (auth luaAuthModule) registerConnectionRequestType(l *lua.LState) {
	mt := l.NewTypeMetatable("ConnectionRequest")
	l.SetGlobal("Person", mt)

	mapping := map[string]lua.LGFunction{
		"ip":       auth.connectionRequestGetIP,
		"port":     auth.connectionRequestGetPort,
		"gatename": auth.connectionRequestGetGatename,
		"key":      auth.connectionRequestGetKey,
	}
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), mapping))
}

func (auth luaAuthModule) connectionRequestGetIP(l *lua.LState) int {
	r, ok := auth.checkConnectionRequest(l)
	if !ok {
		return 0
	}
	l.Push(lua.LString(r.RemoteHost))
	return 1
}

func (auth luaAuthModule) connectionRequestGetPort(l *lua.LState) int {
	r, ok := auth.checkConnectionRequest(l)
	if !ok {
		return 0
	}
	l.Push(lua.LNumber(r.RemotePort))
	return 1
}

func (auth luaAuthModule) connectionRequestGetGatename(l *lua.LState) int {
	r, ok := auth.checkConnectionRequest(l)
	if !ok {
		return 0
	}
	l.Push(lua.LString(r.Gatename))
	return 1
}

func (auth luaAuthModule) connectionRequestGetKey(l *lua.LState) int {
	r, ok := auth.checkConnectionRequest(l)
	if !ok {
		return 0
	}
	l.Push(lua.LString(r.Key))
	return 1
}

func (auth luaAuthModule) checkConnectionRequest(l *lua.LState) (ConnectionRequest, bool) {
	ud := l.CheckUserData(1)
	if r, ok := ud.Value.(ConnectionRequest); ok {
		return r, true
	}
	l.ArgError(1, "Expected ConnectionRequest.")
	return ConnectionRequest{}, false
}
