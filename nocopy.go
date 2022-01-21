package h1

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
