package luajit

import(
    "sync"
    "errors"
)

type GovalueRegistry struct {
    mutex *sync.Mutex    
    registry map[int]interface{}
    currindex int
}

func NewGovalueRegistry() *GovalueRegistry {
    return &GovalueRegistry{
        mutex: &sync.Mutex{}, 
        registry: make(map[int]interface{}),
        currindex: 0,
    }
}

func (this *GovalueRegistry) AddValue(govalue interface{}) int {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    
    this.currindex++    
    this.registry[this.currindex] = govalue
    
    return this.currindex
}

func (this *GovalueRegistry) GetValue(INDEX int) (interface{}, error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    
    val, ok := this.registry[INDEX]
    if !ok {
        return nil, errors.New("Invalid Index Supplied. Index does not exist")
    } else {
        return val, nil
    }    
}

func (this *GovalueRegistry) RemoveValue(INDEX int) error {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    _, ok := this.registry[INDEX]
    if !ok {
        return errors.New("Invalid Index Supplied. Index does not exist")
    } else {
        delete(this.registry, INDEX);
        return nil
    }
}