package luajit

/*
#include <luajit.h>
#include <lua.h>
#include <lualib.h>
*/
import "C"

// Top Lua Constants
const (	
	LUA_MINSTACK  = int(C.LUA_MINSTACK)
	LUA_MULTRET   = int(C.LUA_MULTRET)	
	LUA_YIELD     = int(C.LUA_YIELD)
    LUA_OK        = 0
    LUA_SIGNATURE = string(C.LUA_SIGNATURE)
)

// Top Luajit Constants
const(
    LUAJIT_COPYRIGHT   = string(C.LUAJIT_COPYRIGHT)	
    LUAJIT_VERSION     = string(C.LUAJIT_VERSION)
	LUAJIT_VERSION_NUM = int(C.LUAJIT_VERSION_NUM)
)

// Index constants
const(
	LUA_REGISTRYINDEX = int(C.LUA_REGISTRYINDEX)
	LUA_ENVIRONINDEX  = int(C.LUA_ENVIRONINDEX)
	LUA_GLOBALSINDEX  = int(C.LUA_GLOBALSINDEX)
)

// Error constants
const(
    LUA_ERRERR        = int(C.LUA_ERRERR)
    LUA_ERRMEM        = int(C.LUA_ERRMEM)
    LUA_ERRRUN        = int(C.LUA_ERRRUN)
    LUA_ERRSYNTAX     = int(C.LUA_ERRSYNTAX)
    LUA_ERRUNK_STR    = "UNDEFINED ERROR: "
    LUA_ERRERR_STR    = "ERROR IN ERROR HANDLING: "
    LUA_ERRMEM_STR    = "OUT OF MEMORY ERROR: "
    LUA_ERRRUN_STR    = "RUNTIME ERROR: "
    LUA_ERRSYNTAX_STR = "SYNTAX ERROR: " 	
)

// Type Constants
const (
	LUA_TNONE          = int(C.LUA_TNONE)
	LUA_TNIL           = int(C.LUA_TNIL)
	LUA_TBOOLEAN       = int(C.LUA_TBOOLEAN)
	LUA_TLIGHTUSERDATA = int(C.LUA_TLIGHTUSERDATA)
	LUA_TNUMBER        = int(C.LUA_TNUMBER)
	LUA_TSTRING        = int(C.LUA_TSTRING)
	LUA_TTABLE         = int(C.LUA_TTABLE)
	LUA_TFUNCTION      = int(C.LUA_TFUNCTION)
	LUA_TUSERDATA      = int(C.LUA_TUSERDATA)
	LUA_TTHREAD        = int(C.LUA_TTHREAD)
)