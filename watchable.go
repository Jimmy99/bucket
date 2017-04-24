package bucket

// a basic structure from which to cancel or observe an asynchronous action
type Watchable struct {
	Success chan error

	Cancel  chan error

	Failed  chan error

	// The final observable which the user is likely to read from. Though it can only be fired once it is buffered
	// so that is may be ignored.
	Finished chan error
}

func NewWatchable() *Watchable {
	watchable := &Watchable{ make(chan error), make(chan error), make(chan error), make(chan error, 2) }
	watchable.listen()
	return watchable
}

// Return the intended final observable
func (w *Watchable) Done() chan error {
	return w.Finished
}

// Listen and response to various signals. It will only receive a maximum of one signal by design.
func (w *Watchable) listen() {
	go func(){
		select {
		case err := <- w.Success:
			w.Finished <- err

		case err := <- w.Failed:
			w.Finished <- err
		}

		w.Close()
	}()
}

// Close all channels to prevent memory leaks.
func (w *Watchable) Close(){
	close(w.Success)
	close(w.Cancel)
	close(w.Failed)
	close(w.Finished)
}


