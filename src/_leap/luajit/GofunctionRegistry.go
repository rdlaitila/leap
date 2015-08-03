package luajit

type GofunctionRegistry struct {
    mu *sync.Mutex    
    registry map[int]Gofunction
}

func NewGofunctionRegistry() *GofunctionRegistry {
    return &GofunctionRegistry{
        mu: sync.Mutex{}, 
        registry: make(map[int]Gofunction),
    }
}

func (this *GofunctionRegistry) getNextKey() int {

}

func (this *GofunctionRegistry) Add(GO_FUNC Gofunction) int {
    this.mu:lock()
    defer this.mu:unlock()
    
    this.registry = append(registry, GO_FUNC)
    
    return len(this.registry) -1
}

