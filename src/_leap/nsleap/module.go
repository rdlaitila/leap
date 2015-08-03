package nsleap

import(
    //"log"
    "sync"
    
    "_leap/goluajit"
)

var ModuleMutex *sync.Mutex = &sync.Mutex{}

type Module struct {
}

func NewModule() *Module {
    return &Module{}
}

func (this *Module) Loader(luastate *luajit.State) int { 
    // push module table to stack, this will be returned
    luastate.Newtable() 
    
    // Push nsleap.Mutex
    luastate.Pushfunction(NewMutex)
    luastate.Setfield(-2, "Mutex")
    
    // Push nsleap.WaitGroup
    luastate.Pushfunction(NewWaitGroup)
    luastate.Setfield(-2, "WaitGroup")
    
    // Push nsleap.Thread
    luastate.Pushfunction(NewThread)
    luastate.Setfield(-2, "Thread")
    
    // push module mt to stack
    luastate.Pushmetatable(&luajit.Gometatable{
        IndexFunction: this.index,
    }) 
    
    // Assign module metatable to module table
    luastate.Setmetatable(-2)
    
    return 1
}

func (this *Module) index(ls *luajit.State) int {
    ModuleMutex.Lock()
    defer ModuleMutex.Unlock()
    
    ls.Getfield(-2, ls.Tostring(-1))
    
    return 1
}

/*func (this *Module) newindex(luastate *luajit.State) int {   
    //log.Println("nsleap.__newindex", "numargs:", luastate.Gettop())
    
    luastate.Pushstring("Attempt to add/modify module keys on module 'leap' is disallowed")
    luastate.Error()   
    //luastate.Dostring(`error("Attempt to modify values on module 'leap' is disallowed")`)
    
    return 0 
}*/
