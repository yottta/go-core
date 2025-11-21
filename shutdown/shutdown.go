package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var defaultSigs = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
}

// Wait creates a new chan that will receive items once one of the [defaultSigs] is received.
// [defaultSigs] can be overwritten.
// Once one of the signals is sent to the process, it will be relayed to the channel.
// This method blocks until one signal is received on the channel.
func Wait(overwrite ...os.Signal) {
	signalChan := Chan(overwrite...)
	<-signalChan
}

// Chan creates a new chan that will receive items once one of the [defaultSigs] is received.
// [defaultSigs] can be overwritten.
// Once one of the signals is sent to the process, it will be relayed to the channel allowing
// the client to act on each signal received.
func Chan(overwriteSignals ...os.Signal) <-chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals(overwriteSignals...)...)
	return signalChan
}

// Context returns a [context.Context] that will get cancelled once the process receives one of the signals
// from [defaultSigs]. The signals used to cancel the context can be overwritten by another
// list of [os.Signal] to match the user needs.
// This returns a [context.CancelFunc] that the user is responsible of.
func Context(ctx context.Context, overwriteSignals ...os.Signal) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(ctx, signals(overwriteSignals...)...)
}

func signals(overwrite ...os.Signal) []os.Signal {
	if len(overwrite) > 0 {
		return overwrite
	}
	return defaultSigs
}
