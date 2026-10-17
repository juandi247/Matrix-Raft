package simulator

type SimTime struct {
	Tick uint64
}

func (st *SimTime) Now() uint64 {
	return st.Tick
}
func (st *SimTime) Advance(delta uint64) {
	st.Tick += delta
}
func (st *SimTime) Sleep(uint64) {
}

