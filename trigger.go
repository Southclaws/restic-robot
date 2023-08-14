package main

import "net/http"

// setupTrigger sets up an endpoint for manual triggering of a backup
func (b *backup) setupTrigger() {
	if b.TriggerEndpoint == "" {
		logger.Info("manual trigger disabled")
		return
	}
	http.Handle(b.TriggerEndpoint, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ensure HTTP method is POST
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		logger.Info("manual backup triggered")
		// trigger a backup
		go b.Run()
		// indicate successful trigger attempt
		w.WriteHeader(http.StatusNoContent)
	}))
	logger.Info("manual trigger configured: " + b.TriggerEndpoint)
}
