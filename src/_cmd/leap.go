package main

import(
    "log"
    "os"
    "path/filepath"
    "runtime"
    
    //"_leap/luajit"
    "_leap/goluajit"
    "_leap/nsleap"
)

var boot string = `
leap = require('leap')
threads = {}
`

func main() {
    defer func() {
        if r := recover(); r != nil {
            log.Println(r.(string))
        }
    }()

    //Check for app directory arg
    if len(os.Args) == 1 {
        log.Fatal("No App Directory Specified")
    } else {
        log.Println("OS ARG1:",os.Args[1])
    }

    // Attempt to resolve the app directory
    appdir, abserr := filepath.Abs(os.Args[1]); if abserr != nil {
        log.Fatal("APP PATH:", abserr)
    } 
    log.Println("APP PATH:",appdir)
    
    // Set GOMAXPROCS
    log.Println("MAX PROCS:", runtime.NumCPU())
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Createstate and open libs
    state := luajit.Newstate()    
    state.Openlibs()
    
    // Load modules
    state.Pushmodule("leap", nsleap.NewModule().Loader)
    
    // Load boot.lua
    if bootlerr := state.Loadstring(boot); bootlerr != nil {
        log.Fatal("Error Loading Boot File", bootlerr)
    }
    
    // Call boot.lua
    if bootpcallerr := state.Pcall(0,0,0); bootpcallerr != nil {
        log.Fatal("Error Calling Boot File", bootpcallerr)
    }
    
    //Load in a lua chunk
    if loadfileerr := state.Loadfile(appdir+"/main.lua"); loadfileerr != nil {
        log.Fatal("LOAD MAIN:",loadfileerr)
    }
    
    //Call the lua chunk
    pcallerr := state.Pcall(0,0,0); if pcallerr != nil {
        log.Fatal(pcallerr)        
    }
    
    log.Println("Exiting")
}