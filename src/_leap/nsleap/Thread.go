package nsleap

import(
    "log"
    "sync"
    
    "_leap/goluajit"
    "code.google.com/p/go-uuid/uuid"
)

type Thread struct {
    mu *sync.Mutex
    Gvindex int
}

func NewThread(ls *luajit.State) int {        
    if ls.Gettop() < 1 {
        ls.Pushstring("You must supply a function to leap.Thread() constructor")
        ls.Error()
    }
    
    thread := &Thread{mu: &sync.Mutex{},}
    thread.Gvindex = luajit.Gvregistry.AddValue(thread)
    
    // Create new table. This will be returned
    ls.Newtable()
    
    // Swap arg with table, arg should be at tos
    ls.Insert(ls.Gettop()-1)    
    ls.Setfield(-2, "func")
    
    // Push func
    ls.Pushfunction(thread.run)
    ls.Setfield(-2, "run")
    
    // Create metatable for table
    ls.Pushmetatable(&luajit.Gometatable{
        IndexFunction: thread.index,
        GCFunction: thread.gc,
    })
    
    // Set metatable for table
    ls.Setmetatable(-2)
    
    return 1
}

func (this *Thread) index(ls *luajit.State) int {    
    ls.Getfield(-2, ls.Tostring(-1))
    
    return 1
}

func (this *Thread) run(ls *luajit.State) int {    
    if ls.Gettop() < 1 || ls.Gettop() > 1 {
        panic("Invalid Stack To Thread Run")
    }
    
    threadid := uuid.New()    
    
    ls.Getglobal("threads")    
    threadstate := ls.Newthread()
    ls.Setfield(ls.Gettop()-1, threadid)  
    ls.Pop(1)
    
    ls.Getfield(-1, "func")
    threadstate.Xmove(ls, 1)    
    
    defer func(threadstate *luajit.State) {
        go func(threadstate *luajit.State) {
            if threadstate.Gettop() < 1 || threadstate.Gettop() > 1 {
                panic("Invalid Threadstate Stack")
            }
            //threadstate.Getfield(-2, "func")
            threadstate.Call(0,0)
        }(threadstate)
    }(threadstate)            
    return 0
}

func (this *Thread) gc(ls *luajit.State) int {
    log.Println("THREAD GC")
    return 0
}