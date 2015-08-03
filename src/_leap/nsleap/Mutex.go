package nsleap

import(
    "log"
    "sync"
    
    "_leap/goluajit"    
)

type Mutex struct {
    Ticket chan int
    Gvindex int
    mu *sync.Mutex
}

func NewMutex(ls *luajit.State) int {
    luajit.GlobalMutex.Lock()
    defer luajit.GlobalMutex.Unlock()
    
    mu := &Mutex{
        Ticket: make(chan int, 1),
        mu: &sync.Mutex{},
    }
    mu.Ticket <- 1
    
    // Obtain a gvindex 
    mu.Gvindex = luajit.Gvregistry.AddValue(mu)
    
    // Create new userdata. This will be returned
    ls.Newtable()
    
    // Push mutex.Lock
    ls.Pushfunction(mu.lock)
    ls.Setfield(-2, "lock")
    
    // Push mutex.Unlock
    ls.Pushfunction(mu.unlock)
    ls.Setfield(-2, "unlock")
    
    // Create metatable for userdat
    ls.Pushmetatable(&luajit.Gometatable{
        IndexFunction: mu.index,
        GCFunction: mu.gc,
    })
    
    // Set metatable for userdata
    ls.Setmetatable(-2,)
    
    return 1
}

func (this *Mutex) index(ls *luajit.State) int {    
    ls.Getfield(-2, ls.Tostring(-1))
    
    return 1
}

func (this *Mutex) gc(ls *luajit.State) int {
    log.Println("MUTEX GC")
    luajit.Gvregistry.RemoveValue(this.Gvindex)
    return 0
}

func (this *Mutex) lock(ls *luajit.State) int {    
    <- this.Ticket
    return 0
}

func (this *Mutex) unlock(ls *luajit.State) int {    
    this.Ticket <- 1
    return 0
}