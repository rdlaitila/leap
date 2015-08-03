package luajit

type Gometatable struct {
    IndexFunction    Gofunction
    NewindexFunction Gofunction
    TostringFunction Gofunction    
    GCFunction       Gofunction
}

func (this *Gometatable) Index() Gofunction {
    return this.IndexFunction
}

func (this *Gometatable) Newindex() Gofunction {
    return this.NewindexFunction
}

func (this *Gometatable) Tostring() Gofunction {
    return this.TostringFunction
}

func (this *Gometatable) GC() Gofunction {
    return this.GCFunction
}