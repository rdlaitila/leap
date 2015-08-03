package nsleap

import(
    "sync"
    "log"
    "_leap/goluajit"
)

type WaitGroup struct{
    Gvindex int
    wg *sync.WaitGroup
    mu *sync.Mutex
}

func NewWaitGroup(ls *luajit.State) int {
    luajit.GlobalMutex.Lock()
    defer luajit.GlobalMutex.Unlock()

    wg := &WaitGroup{wg: &sync.WaitGroup{}, mu: &sync.Mutex{}}
    
    // Obtain a gvindex 
    wg.Gvindex = luajit.Gvregistry.AddValue(wg)
    
    // Create new table. This will be returned
    ls.Newtable()
    
    // Push add
    ls.Pushfunction(wg.add)
    ls.Setfield(-2, "add")
    
    // Push done
    ls.Pushfunction(wg.done)
    ls.Setfield(-2, "done")
    
    // Push wait
    ls.Pushfunction(wg.wait)
    ls.Setfield(-2, "wait")
    
    // Create metatable for table
    ls.Pushmetatable(&luajit.Gometatable{
        GCFunction: wg.gc,
    })
    
    // Set metatable for userdata
    ls.Setmetatable(-2,)
    
    return 1
}

func (this *WaitGroup) add(ls *luajit.State) int {
    luajit.GlobalMutex.Lock()
    defer luajit.GlobalMutex.Unlock()
    
    this.wg.Add(1)
    
    return 0
}

func (this *WaitGroup) done(ls *luajit.State) int {
    luajit.GlobalMutex.Lock()
    defer luajit.GlobalMutex.Unlock()
    
    this.wg.Done()
    
    return 0
}

func (this *WaitGroup) wait(ls *luajit.State) int {    
    this.wg.Wait()
    
    return 0
}

func (this *WaitGroup) gc(ls *luajit.State) int {
    log.Println("WAITGROUP GC")
    return 0
}