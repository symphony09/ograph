package coro

func New[In, Out any](f func(in In, yield func(Out) In) Out) (resume func(In) (Out, bool)) {
	cin := make(chan In)
	cout := make(chan Out)
	running := true
	resume = func(in In) (out Out, ok bool) {
		if !running {
			return
		}
		cin <- in
		out = <-cout
		return out, running
	}
	yield := func(out Out) In {
		cout <- out
		return <-cin
	}
	go func() {
		out := f(<-cin, yield)
		running = false
		cout <- out
	}()
	return resume
}
