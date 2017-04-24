package bucket

type Watchable struct {
	// signal indicating an action is completed
	Success chan error

	// a channel from which to cancel an asynchronous action
	Cancel  chan error

	// a read only channel from which to receive errors
	Failed  chan error

	// end the Done() loop to prevent any memory leaks
	Finished chan error
}

func NewWatchable() *Watchable {
	return &Watchable{ make(chan error), make(chan error), make(chan error), make(chan error) }
}

// Return a channel that will only be fired a maximum of one time from a Watchable event
func (w *Watchable) Done() chan error {
	go func(){
		select {
		case err := <- w.Success:
			w.Finished <- err

		case err := <- w.Failed:
			w.Finished <- err
		}

		w.Close()
	}()

	return w.Finished
}

// Close all channels to prevent memory leaks
func (w *Watchable) Close(){
	close(w.Success)
	close(w.Cancel)
	close(w.Failed)
	close(w.Finished)
}


