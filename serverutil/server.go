package serverutil

// StopFunc is a server stop function, typically returned from Serve()
type StopFunc func()

// Server represents a component that serves stuff, and can be gracefully stopped.
type Server interface {
	Serve() StopFunc
}

// Runner is used to stop/start multiple servers
type Runner struct {
	servers  []Server
	stoppers []StopFunc
}

// RunAll will run all given servers and return a Runner instance
func RunAll(servers ...Server) *Runner {
	r := &Runner{}
	stoppers := make([]StopFunc, len(servers))
	for i, server := range servers {
		stoppers[i] = server.Serve()
	}
	r.servers = servers
	r.stoppers = stoppers
	return r
}

// StopFunc will stop all registered servers in reverse order from how they were registered
func (r *Runner) StopFunc() {
	for index := len(r.stoppers) - 1; index >= 0; index-- {
		r.stoppers[index]()
	}
}
