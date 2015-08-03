package luajit

import(
    "errors"
    "unsafe"
    "fmt"
)

/*
#include <luajit.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include <stddef.h>
#include <stdlib.h>

extern void goluajit_luainit(lua_State*);
extern void goluajit_pushclosure(lua_State*, int);
*/
import "C"

// State is a golang struct wrapping our C lua_State
type State struct {
    luastate *C.lua_State
    gvindex int
}

// NewState Creates a new Lua state. It calls luaL_newstate which calls lua_newstate with an allocator based 
// on the standard C realloc function and then sets a panic function (see lua_atpanic) 
// that prints an error message to the standard error output in case of fatal errors. 
// TODO: Handle NULL return of luaL_newstate and error appropriately
// TODO: Set panic function for lua_atpanic
func Newstate() *State {
    state := &State{
        luastate: C.luaL_newstate(),
    }
    state.Init()
    return state
    
}
// geterror attempts to find the lua error type constant and error string
// from the stack, and generate a go error
func (this *State) geterror(errno int) error {
    if errno == 0 {
        return nil
    }
        
    errstr := ""
        
    switch errno {
        case LUA_ERRERR:
            errstr = LUA_ERRERR_STR
            break
        case LUA_ERRSYNTAX:
            errstr = LUA_ERRSYNTAX_STR
            break
        case LUA_ERRRUN:
            errstr = LUA_ERRRUN_STR
            break
        case LUA_ERRMEM:
            errstr = LUA_ERRMEM_STR
            break
        default:
            errstr = LUA_ERRUNK_STR
    }
    
    if this.Gettop() == 0 {
        return errors.New(errstr + "No Error Available On Stack")
    } else if this.Type(-1) == LUA_TSTRING {
        return errors.New(errstr + this.Tostring(-1))
    } else {
        return errors.New(errstr + "Unable to locate error on stack")
    }
}

//export docallback
func docallback(stateindex C.float, funcindex C.float) int {
    // pull our *State value and Gofunction value from GovalueRegistry
    stateval, stateerr := Gvregistry.GetValue(int(stateindex)); if stateerr != nil {
        panic(stateerr.Error())
    }
    fnval, fnvalerr := Gvregistry.GetValue(int(funcindex)); if fnvalerr != nil {
        panic(fnvalerr.Error())
    }
    
    // Cast to their proper types
    state, ok := stateval.(*State); if !ok {
        panic("Error Casting State Interface")
    }
	fn, ok := fnval.(Gofunction); if !ok {
        panic("Error Casting Gofunction Interface")
    }
    
    //Call function passing state
    return fn(state)
}

// Init configures internal values of the luajit.State object. This is called
// automatically by luajit.Newstate() and should only be called if the State struct
// was instantiated manually.
func (this *State) Init() {
    this.gvindex = Gvregistry.AddValue(this)
    C.goluajit_luainit(this.luastate)
}

// Yields a coroutine.
//
// This function should only be called as the return expression of a Go
// function, as follows:
// 	return s.Yield(nresults)
//
// When a Go function calls Yield in that way, the running coroutine
// suspends its execution, and the call to Resume that started this coroutine
// returns. The parameter nresults is the number of values from the stack
// that are passed as results to Resume.
func (this *State) Yield(nresults int) int {
	return int(C.lua_yield(this.luastate, C.int(nresults)))
}

// Exchange values between different threads of the /same/ global state.
//
// This function pops n values from the stack from, and pushes them onto
// the stack to.
func (this *State) Xmove(from *State, n int) {
	C.lua_xmove(from.luastate, this.luastate, C.int(n))
}

//TODO: lua_upvalueindex

// Returns the name of the type encoded by the value tp, which must be one
// the values returned by Type.
func (this *State) Typename(tp int) string {
    //return C.GoString(C.lua_typename(this.luastate, C.int(tp)))
    
    typename := ""
    
    switch this.Type(tp) {
        case LUA_TTABLE:
            typename = "table"
            break
        case LUA_TSTRING:
            typename = "string"
            break
        case LUA_TNUMBER:
            typename = "number"
            break
        case LUA_TBOOLEAN:
            typename = "boolean"
            break
        case LUA_TFUNCTION:
            typename = "function"
            break;
        case LUA_TNIL:
            typename = "nil"
            break
        case LUA_TNONE:
            typename = "none"
            break
        case LUA_TUSERDATA:
            typename = "userdata"
            break
        case LUA_TLIGHTUSERDATA:
            typename = "lightuserdata"
            break
        case LUA_TTHREAD:
            typename = "thread"
            break
        default:
            typename = "unknown"
            break
    }
    
    return typename
}

// Returns the type of the value in the given valid index, or luajit.LUA_TNONE 
// for a non-valid index (that is, an index to an "empty" stack position). The
// types returned by lua_type are coded by the following constants defined in
// const.go: LUA_TNIL, LUA_TNUMBER, LUA_TBOOLEAN, LUA_TSTRING, LUA_TTABLE, 
// LUA_TFUNCTION, LUA_TUSERDATA, LUA_TTHREAD, and LUA_TLIGHTUSERDATA.
func (this *State) Type(index int) int {
	return int(C.lua_type(this.luastate, C.int(index)))
}

// If the value at the given valid index is a full userdata, returns
// its block address. If the value is a light userdata, returns its
// pointer. Otherwise, returns unsafe.Pointer(nil).
func (this *State) Touserdata(index int) unsafe.Pointer {
	return C.lua_touserdata(this.luastate, C.int(index))
}

// Converts the value at the given valid index to a Lua thread
// (represented as a *State). This value must be a thread; otherwise,
// the function returns nil.
func (this *State) Tothread(index int) *State {
	t := C.lua_tothread(this.luastate, C.int(index))
	if t == nil {
		return nil
	}
	return &State{luastate: t}
}

// Converts the Lua value at the given valid index to a Go
// string. The Lua value must be a string or a number; otherwise,
// the function returns an empty string. If the value is a number, then
// Tostring also changes the actual value in the stack to a string. (This
// change confuses Next when Tostring is applied to keys during a table
// traversal).  The string always has a zero ('\0') after its last
// character (as in C), but can contain other zeros in its body.
func (this *State) Tostring(index int) string {
	str := C.lua_tolstring(this.luastate, C.int(index), nil)
	if str == nil {
		return ""
	}
	return C.GoString(str)
}

// Converts the value at the given acceptable index to a uintptr. The
// value can be a userdata, a table, a thread, or a function; otherwise,
// Topointer returns nil. Different objects will give different
// pointers. There is no way to convert the pointer back to its original
// value.
//
// Typically this function is used only for debug information.
func (this *State) Topointer(index int) unsafe.Pointer {
	return C.lua_topointer(this.luastate, C.int(index))
}

// Converts the Lua value at the given valid index to a float64. The
// Lua value must be a number or a string convertible to a number; otherwise,
// Tonumber returns 0.
func (this *State) Tonumber(index int) float64 {
	return float64(C.lua_tonumber(this.luastate, C.int(index)))
}

//TODO: lua_tolstring

// Converts the Lua value at the given valid index to a Go int. The Lua
// value must be a number or a string convertible to a number; otherwise,
// Tointeger returns 0.
//
// If the number is not an integer, it is truncated in some non-specified
// way.
func (this *State) Tointeger(index int) int {
	return int(C.lua_tointeger(this.luastate, C.int(index)))
}

//TODO: lua_tocfunction

// Converts the Lua value at the given valid index to a Go boolean
// value. Like all tests in Lua, Toboolean returns true for any Lua value
// different from false and nil; otherwise it returns false. It also returns
// false when called with a non-valid index. (If you want to accept only
// actual boolean values, use Isboolean to test the value's type.)
func (this *State) Toboolean(index int) bool {
	return int(C.lua_toboolean(this.luastate, C.int(index))) == 1
}

// Returns the status of the thread s.
//
// The status can be 0 for a normal thread, an error code if the thread
// finished its execution with an error, or luajit.Yield if the thread
// is suspended.
func (this *State) Status() int {
	return int(C.lua_status(this.luastate))
}

// Sets the value of a closure's upvalue. It assigns the value at the top
// of the stack to the upvalue and returns its name. It also pops the value
// from the stack. Parameters funcindex and n are as in the Getupvalue.
//
// Returns an error (and pops nothing) when the index is greater
// than the number of upvalues.
func (this *State) Setupvalue(funcindex, n int) (string, error) {
	r := C.lua_setupvalue(this.luastate, C.int(funcindex), C.int(n))
	if r == nil {
		return "", errors.New("index exceeds number of upvalues")
	}
	return C.GoString(r), nil
}

// Accepts any valid index, or 0, and sets the stack top to this
// index. If the new top is larger than the old one, then the new elements
// are filled with nil. If index is 0, then all stack elements are removed.
func (this *State) Settop(index int) {
	C.lua_settop(this.luastate, C.int(index))
}

// Does the equivalent to t[k] = v, where t is the value at the given valid
// index, v is the value at the top of the stack, and k is the value just
// below the top.
//
// This function pops both the key and the value from the stack. As in Lua,
// this function may trigger a metamethod for the "newindex" event.
func (this *State) Settable(index int) {
	C.lua_settable(this.luastate, C.int(index))
}

// Pops a table from the stack and sets it as the new metatable for the
// value at the given valid index.
func (this *State) Setmetatable(index int) int {
	return int(C.lua_setmetatable(this.luastate, C.int(index)))
}

//TODO: lua_setlocal
//TODO: lua_sethook

// Pops a value from the stack and sets it as the new value of global name.
func (this *State) Setglobal(name string) {
	this.Setfield(LUA_GLOBALSINDEX, name)
}

// Does the equivalent to t[k] = v, where t is the value at the given valid
// index and v is the value at the top of the stack.
//
// This function pops the value from the stack. As in Lua, this function
// may trigger a metamethod for the "newindex" event
func (this *State) Setfield(index int, k string) {
	ck := C.CString(k)
	defer C.free(unsafe.Pointer(ck))
	C.lua_setfield(this.luastate, C.int(index), ck)
}

//TODO: lua_setfenv
//TODO: lua_setallocf

// Starts and resumes a coroutine in a given thread.
//
// To start a coroutine, you first create a new thread (see Newthread);
// then you push onto its stack the main function plus any arguments; then
// you call Resume, with narg being the number of arguments. This call
// returns when the coroutine suspends or finishes its execution. When
// it returns, the stack contains all values passed to Yield, or all
// values returned by the body function. Resume returns (true, nil) if the
// coroutine yields, (false, nil) if the coroutine finishes its execution
// without errors, or (false, error) in case of errors (see Pcall).
//
// In case of errors, the stack is not unwound, so you can use the debug
// API over it. The error message is on the top of the stack.
//
// To resume a coroutine, you remove any results from the last Yield,
// put on its stack only the values to be passed as results from the yield,
// and then call Resume.
func (this *State) Resume(narg int) (yield bool, e error) {
	switch r := int(C.lua_resume(this.luastate, C.int(narg))); {
	case r == LUA_YIELD:
		return true, nil
	case r == LUA_OK:
		return false, nil
	default:
		return false, this.geterror(r)
	}
}

// Moves the top element into the given position (and pops it), without
// shifting any element (therefore replacing the value at the given
// position).
func (this *State) Replace(index int) {
	C.lua_replace(this.luastate, C.int(index))
}

// Removes the element at the given valid index, shifting down the elements
// above this index to fill the gap. Cannot be called with a pseudo-index,
// because a pseudo-index is not an actual stack position.
func (this *State) Remove(index int) {
	C.lua_remove(this.luastate, C.int(index))
}

// Sets the Go function fn as the new value of global name.
func (this *State) Register(fn Gofunction, name string) {
	this.Pushclosure(fn, 0)
	this.Setglobal(name)
}

// Does the equivalent of t[n] = v, where t is the value at the given valid
// index and v is the value at the top of the stack.
//
// This function pops the value from the stack. The assignment is raw;
// that is, it does not invoke metamethods.
func (this *State) Rawseti(index, n int) {
	C.lua_rawseti(this.luastate, C.int(index), C.int(n))
}

// Similar to Settable, but does a raw assignment (i.e., without
// metamethods).
func (this *State) Rawset(index int) {
	C.lua_rawset(this.luastate, C.int(index))
}

// Pushes onto the stack the value t[n], where t is the value at the given
// valid index. The access is raw; that is, it does not invoke metamethods.
func (this *State) Rawgeti(index, n int) {
	C.lua_rawgeti(this.luastate, C.int(index), C.int(n))
}

// Similar to Gettable, but does a raw access (i.e., without metamethods).
func (this *State) Rawget(index int) {
	C.lua_rawget(this.luastate, C.int(index))
}

// Returns true if the two values at valid indices i1 and i2 are
// primitively equal (that is, without calling metamethods). Otherwise
// returns false. Also returns false if any of the indices are invalid.
func (this *State) Rawequal(i1, i2 int) bool {
	return int(C.lua_rawequal(this.luastate, C.int(i1), C.int(i2))) == 1
}

//TODO: lua_pushvfstring

// Pushes a copy of the element at the given valid index onto the stack.
func (this *State) Pushvalue(index int) {
	C.lua_pushvalue(this.luastate, C.int(index))
}

// Pushes the thread represented by s onto the stack. Returns 1 if this
// thread is the main thread of its state.
func (this *State) Pushthread() int {
	return int(C.lua_pushthread(this.luastate))
}

// Pushes the string str onto the stack.
func (this *State) Pushstring(str string) {
	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	C.lua_pushstring(this.luastate, cs)
}

// Pushes a number with value n onto the stack.
func (this *State) Pushnumber(n float64) {
	C.lua_pushnumber(this.luastate, C.lua_Number(n))
}

//TODO: lua_pushnil
//TODO: lua_pushlstring
//TODO: lua_pushliteral
//TODO: lua_pushlightuserdata
//TODO: lua_pushinteger
//TODO: lua_pushfstring
//TODO: lua_pushcfunction

// Pushes a new Go closure onto the stack.
//
// When a Go function is created, it is possible to associate some
// values with it, thus creating a Go closure; these values are then
// accessible to the function whenever it is called. To associate values
// with a Go function, first these values should be pushed onto the stack
// (when there are multiple values, the first value is pushed first). Then
// Pushclosure is called to create and push the Go function onto the
// stack, with the argument n telling how many values should be associated
// with the function. Pushclosure also pops these values from the stack.
//
// The maximum value for n is 254.
func (this *State) Pushclosure(fn Gofunction, n int) {
    if !this.Checkstack(1) {
        panic("STATE: unable to grow lua_state stack")
    }

    C.lua_pushnumber(this.luastate, C.lua_Number(C.float(float64(this.gvindex))))
    C.lua_pushnumber(this.luastate, C.lua_Number(C.float(float64(Gvregistry.AddValue(fn)))))
	C.goluajit_pushclosure(this.luastate, C.int(n + 2))
}

// Pushes a Go function onto the stack. This function receives a pointer to
// a Go function and pushes onto the stack a Lua value of type function that,
// when called, invokes the corresponding Go function.
//
// Any function to be registered in Lua must follow the correct protocol
// to receive its parameters and return its results (see Gofunction).
func (this *State) Pushfunction(fn Gofunction) {
    this.Pushclosure(fn, 0)
}

// Pushmodule pushes a module loader function into package.preload, permitting you to invoke 
// a go function when conducting lua requires ex:  "require('your_module_name')"
func (this *State) Pushmodule(name string, fn Gofunction) {    
    this.Getglobal("package")
    this.Getfield(-1, "preload")
    this.Pushfunction(fn)
    this.Setfield(-2, name)
    this.Pop(2)
}

// Pushmetatable pushes a Gometatable struct to the stack and conducts the appropriate mapping
// of metatable keys to Gofunctions.
// 
// you must still use Setmetatable after this function returns to assign your metatable to some 
// other table on the stack.
func (this *State) Pushmetatable(mt *Gometatable) {
    this.Newtable()    
    
    if mt.Index() != nil {
        this.Pushfunction(mt.Index())
        this.Setfield(-2, "__index")
    }
    if mt.Newindex() != nil {
        this.Pushfunction(mt.Newindex())
        this.Setfield(-2, "__newindex")
    }
    if mt.Tostring() != nil {
        this.Pushfunction(mt.Tostring())
        this.Setfield(-2, "__tostring")
    }
    if mt.GC() != nil {
        this.Pushfunction(mt.GC())
        this.Setfield(-2, "__gc")
    }
}

//TODO: lua_pushboolean

// Pops n elements from the stack.
func (this *State) Pop(index int) {
	this.Settop(-index - 1)
}

// Calls a function in protected mode.
//
// Both nargs and nresults have the same meaning as in Call. If there are
// no errors during the call, Pcall behaves exactly like Call. However,
// if there is any error, Pcall catches it, pushes a single value on the
// stack (the error message), and returns an error code. Like Call, Pcall
// always removes the function and its arguments from the stack.
//
// If errfunc is 0, then the error message returned on the stack is exactly
// the original error message. Otherwise, errfunc is the stack index of
// an error handler function. (In the current implementation, this index
// cannot be a pseudo-index.) In case of runtime errors, this function
// will be called with the error message and its return value will be the
// message returned on the stack by Pcall.
//
// Typically, the error handler function is used to add more debug
// information to the error message, such as a stack traceback. Such
// information cannot be gathered after the return of Pcall, since by then
// the stack has unwound.
func (this *State) Pcall(nargs, nresults, errfunc int) error {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println(r.(string))
        }
    }()
        
    r := int(C.lua_pcall(this.luastate, C.int(nargs), C.int(nresults), C.int(errfunc)))        
    return this.geterror(r)
}

//TODO: lua_objlen
//TODO: lua_next

// Newuserdata.  This function allocates a new block of memory with the given size, 
// pushes onto the stack a new full userdata with the block address, and returns this 
// address.
//
// Userdata represent C values in Lua. A full userdata represents a block of memory. It
// is an object (like a table): you must create it, it can have its own metatable, and 
// you can detect when it is being collected. A full userdata is only equal to itself (under raw equality).
//
// When Lua collects a full userdata with a gc metamethod, Lua calls the metamethod and marks 
// the userdata as finalized. When this userdata is collected again then Lua frees its corresponding memory. 
func (this *State) Newuserdata() {
    C.lua_newuserdata(this.luastate, C.size_t(1))
}

// Creates a new thread, pushes it on the stack, and returns a pointer
// to a State that represents this new thread. The new state returned by
// this function shares with the original state all global objects (such
// as tables), but has an independent execution stack.
//
// There is no explicit function to close or to destroy a thread. Threads
// are subject to garbage collection, like any Lua object.
func (this *State) Newthread() *State {
    if !this.Checkstack(1) {
        panic("STATE: unable to grow lua_state stack")
    }
    
    newstate := &State{
        luastate: C.lua_newthread(this.luastate),
    }
    
	return newstate
}

// Creates a new empty table and pushes it onto the stack. It is equivalent
// to Createtable(0, 0).
func (this *State) Newtable() {
	this.Createtable(0, 0)
}

//TODO: lua_newstate
//TODO: lua_load
//TODO: lua_lessthan

// Returns true if the value at the given acceptable index is a userdata
// (either full or light), and false otherwise.
func (this *State) Isuserdata(index int) bool {
	t := this.Type(index)
	return t == LUA_TUSERDATA || t == LUA_TLIGHTUSERDATA
}

// Returns true if the value at the given valid index is a thread,
// and false otherwise.
func (this *State) Isthread(index int) bool {
	return this.Type(index) == LUA_TTHREAD
}

// Returns true if the value at the given valid index is a table,
// and false otherwise.
func (this *State) Istable(index int) bool {
	return this.Type(index) == LUA_TTABLE
}

// Returns true if the value at the given valid index is a string,
// and false otherwise.
func (this *State) Isstring(index int) bool {
	return this.Type(index) == LUA_TSTRING
}

// Returns true if the value at the given valid index is a number,
// and false otherwise.
func (this *State) Isnumber(index int) bool {
	return this.Type(index) == LUA_TNUMBER
}

// Returns true if the given valid index is not valid (that is, it
// refers to an element outside the current stack) or if the value at this
// index is nil, and false otherwise.
func (this *State) Isnoneornil(index int) bool {
	return this.Type(index) <= 0
}

// Returns true if the given valid index is not valid (that is, it
// refers to an element outside the current stack), and false otherwise.
func (this *State) Isnone(index int) bool {
	return this.Type(index) == LUA_TNONE
}

// Returns true if the value at the given valid index is nil,
// and false otherwise.
func (this *State) Isnil(index int) bool {
	return this.Type(index) == LUA_TNIL
}

// Returns true if the value at the given valid index is light
// userdata, and false otherwise.
func (this *State) Islightuserdata(index int) bool {
	return this.Type(index) == LUA_TLIGHTUSERDATA
}

// Returns true if the value at the given valid index is a function
// (either Go or Lua), and false otherwise.
func (this *State) Isfunction(index int) bool {
	return this.Type(index) == LUA_TFUNCTION
}

// Returns true if the value at the given valid index is a Go function,
// and false otherwise.
func (this *State) Isgofunction(index int) bool {
	return int(C.lua_iscfunction(this.luastate, C.int(index))) == 1
}

// Returns true if the value at the given valid index has type
// boolean, and false otherwise.
func (this *State) Isboolean(index int) bool {
	return this.Type(index) == LUA_TBOOLEAN
}

// Moves the top element into the given valid index, shifting up the elements
// above this index to open space. Cannot be called with a pseudo-index,
// because a pseudo-index is not an actual stack position.
func (this *State) Insert(index int) {
	C.lua_insert(this.luastate, C.int(index))
}

//TODO: lua_getupvalue

// Returns the index of the top element in the stack. Because indices start
// at 1, this result is equal to the number of elements in the stack (and
// so 0 means an empty stack).
func (this *State) Gettop() int {    
	return int(C.lua_gettop(this.luastate))
}

// Pushes onto the stack the value t[k], where t is the value at the
// given valid index and k is the value at the top of the stack.
//
// This function pops the key from the stack (putting the resulting value
// in its place). As in Lua, this function may trigger a metamethod for
// the "index" event
func (this *State) Gettable(index int) {
	C.lua_gettable(this.luastate, C.int(index))
}

//TODO: lua_getstack
//TODO: lua_getmetatable
//TODO: lua_getlocal
//TODO: lua_getinfo
//TODO: lua_gethookmask
//TODO: lua_gethookcount
//TODO: lua_gethook

// Pushes onto the stack the value of the global name.
func (this *State) Getglobal(name string) {
	this.Getfield(LUA_GLOBALSINDEX, name)
}

// Pushes onto the stack the value t[k], where t is the value at the
// given valid index.
func (this *State) Getfield(index int, k string) {
	cs := C.CString(k)
	defer C.free(unsafe.Pointer(cs))
	C.lua_getfield(this.luastate, C.int(index), cs)
}

//TODO: lua_getfenv
//TODO: lua_getallocf
//TODO: lua_gc

// Generates a Lua error. The error message (which can actually be a Lua
// value of any type) must be on the stack top. This function does a long
// jump, and therefore never returns.
func (this *State) Error() {
    C.lua_error(this.luastate)
    
    /*err := this.Tostring(-1)    
    fmt.Println("ERROR: ", err)
    this.Dostring(`print(debug.traceback())`)    
    panic("goluajit Panic")*/
}

//TODO: lua_equal
//TODO: lua_dump

// Creates a new empty table and pushes it onto the stack. The new table
// has space pre-allocated for narr array elements and nrec non-array
// elements. This pre-allocation is useful when you know exactly how many
// elements the table will have. Otherwise you can use the function Newtable.
func (this *State) Createtable(narr, nrec int) {
    if !this.Checkstack(1) {
        panic("STATE: unable to grow lua_state stack")
    }
	C.lua_createtable(this.luastate, C.int(narr), C.int(nrec))
}

//TODO: lua_cpcall

// Concatenates the n values at the top of the stack, pops them, and
// leaves the result at the top. If n is 1, the result is the single
// value on the stack (that is, the function does nothing); if n is 0,
// the result is the empty string. Concatenation is performed following
// the usual semantics of Lua.
func (this *State) Concat(n int) {    
	C.lua_concat(this.luastate, C.int(n))
}

// Destroys all objects in the given Lua state (calling the corresponding
// garbage-collection metamethods, if any) and frees all dynamic memory
// used by this state. On several platforms, you may not need to call
// this function, because all resources are naturally released when the
// host program ends. On the other hand, long-running programs, such as
// a daemon or a web server, might need to release states as soon as they
// are not needed, to avoid growing too large.
func (this *State) Close() {
	C.lua_close(this.luastate)
}

// Ensures that there are at least extra free stack slots in the stack. It
// returns false if it cannot grow the stack to that size. This function
// never shrinks the stack; if the stack is already larger than the new
// size, it is left unchanged.
func (this *State) Checkstack(extra int) bool {
	return int(C.lua_checkstack(this.luastate, C.int(extra))) == 1
}

// Calls a function.
//
// To call a function you must use the following protocol: first,
// the function to be called is pushed onto the stack; then, the
// arguments to the function are pushed in direct order; that is, the
// first argument is pushed first. Finally you call Call; nargs is the
// number of arguments that you pushed onto the stack. All arguments
// and the function value are popped from the stack when the function
// is called. The function results are pushed onto the stack when the
// function returns. The number of results is adjusted to nresults, unless
// nresults is luajit.Multret. In this case, all results from the function
// are pushed. Lua takes care that the returned values fit into the stack
// space. The function results are pushed onto the stack in direct order
// (the first result is pushed first), so that after the call the last
// result is on the top of the stack.
//
// Any error inside the called function is propagated upwards (with
// a longjmp).
func (this *State) Call(nargs, nresults int) {
	C.lua_call(this.luastate, C.int(nargs), C.int(nresults))
}

//TODO: lua_atpanic. NOTE: not sure if this is possible to pass back to golang. 
// use pcall instead on first script run and pass an index of a function on the stack.

//TODO: lua_Writer
//TODO: lua_State
//TODO: lua_Reader
//TODO: lua_Number
//TODO: lua_Integer
//TODO: lua_Hook
//TODO: lua_Debug
//TODO: lua_CFunction
//TODO: lua_Alloc
//TODO: luaL_where
//TODO: luaL_unref
//TODO: luaL_typerror
//TODO: luaL_register
//TODO: luaL_ref
//TODO: luaL_pushresult
//TODO: luaL_prepbuffer
//TODO: luaL_optstring
//TODO: luaL_optnumber
//TODO: luaL_optlstring
//TODO: luaL_optlong
//TODO: luaL_optinteger
//TODO: luaL_optint

// Openlibs Opens all standard Lua libraries into the given state. 
// http://www.lua.org/manual/5.1/manual.html#luaL_openlibs
func (this *State) Openlibs() {
    C.luaL_openlibs(this.luastate)
}

//TODO: luaL_newmetatable

// Loads a string as a Lua chunk.
//
// This function only loads the chunk; it does not run it.
func (this *State) Loadstring(str string) error {
	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
    
	r := int(C.luaL_loadstring(this.luastate, cs))
    
	return this.geterror(r)
}

// Loadfile Loads a file as a Lua chunk. This function uses lua_load to load 
// the chunk in the file named filename. The first line in the file is ignored 
// if it starts with a #
//
// As lua_load, this function only loads the chunk; it does not run it. 
//
// http://www.lua.org/manual/5.1/manual.html#luaL_loadfile
func (this *State) Loadfile(filename string) error {
    cs := C.CString(filename)
    defer C.free(unsafe.Pointer(cs))
    
    return this.geterror(int(C.luaL_loadfile(this.luastate, cs)))
}

//TODO: luaL_loadbuffer
//TODO: luaL_gsub
//TODO: luaL_getmetatable
//TODO: luaL_getmetafield
//TODO: luaL_error

// Dostring Loads and runs the given string. It returns 0 if there are no errors or 1 in case of errors.
func (this *State) Dostring(str string) int {
    loadret := this.Loadstring(str)    
    if loadret != nil {
        return 1
    }

    pcallret := this.Pcall(0,LUA_MULTRET,0)
    if pcallret != nil {
        return 1
    }
    
    return 0
}

// Dofile Loads and runs the given file. It returns 0 if there are no errors or 1 in case of errors. 
func (this *State) Dofile(path string) int {
    loadret := this.Loadfile(path)    
    if loadret != nil {
        return 1
    }

    pcallret := this.Pcall(0,LUA_MULTRET,0)
    if pcallret != nil {
        return 1
    }
    
    return 0
}

//TODO: luaL_checkudata
//TODO: luaL_checktype
//TODO: luaL_checkstring
//TODO: luaL_checkstack
//TODO: luaL_checkoption
//TODO: luaL_checknumber
//TODO: luaL_checklstring
//TODO: luaL_checklong
//TODO: luaL_checkinteger
//TODO: luaL_checkint
//TODO: luaL_checkany
//TODO: luaL_callmeta
//TODO: luaL_buffinit
//TODO: luaL_argerror
//TODO: luaL_argcheck
//TODO: luaL_addvalue
//TODO: luaL_addstring
//TODO: luaL_addsize
//TODO: luaL_addlstring
//TODO: luaL_addchar
//TODO: luaL_Reg
//TODO: luaL_Buffer