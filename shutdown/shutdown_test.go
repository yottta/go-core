package shutdown

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const (
	envKeyForShutdown = "shutdown_method"

	shutdownMethodWait    = "wait"
	shutdownMethodChan    = "chan"
	shutdownMethodContext = "context"
)

func TestMain(t *testing.M) {
	if method, ok := os.LookupEnv(envKeyForShutdown); ok {
		res := result{
			startedAt: time.Now(),
		}
		switch method {
		case shutdownMethodWait:
			Wait()
			res.executedMethod = method // writing it here to be sure that this is written only when the shutdown method is actually executed
		case shutdownMethodChan:
			<-Chan()
			res.executedMethod = method // writing it here to be sure that this is written only when the shutdown method is actually executed
		case shutdownMethodContext:
			ctx, cancel := Context(context.Background())
			defer cancel()
			<-ctx.Done()
			res.executedMethod = method // writing it here to be sure that this is written only when the shutdown method is actually executed
		default:
			fmt.Println("invalid shutdown method provided")
			os.Exit(2)
		}
		res.stoppedAt = time.Now()
		fmt.Printf("%s", res.encode())
		os.Exit(0)
	}
	os.Exit(t.Run())
}

func TestShutdownMethods(t *testing.T) {
	cases := map[string]struct {
		delayBeforeSendingSignal time.Duration
		signalToSend             syscall.Signal
		shutdownMethod           string
	}{
		"wait - send SIGINT after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodWait,
		},
		"wait - send SIGTERM after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodWait,
		},
		"wait - send SIGINT after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodWait,
		},
		"wait - send SIGTERM after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodWait,
		},
		"chan - send SIGINT after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodChan,
		},
		"chan - send SIGTERM after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodChan,
		},
		"chan - send SIGINT after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodChan,
		},
		"chan - send SIGTERM after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodChan,
		},
		"context - send SIGINT after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodContext,
		},
		"context - send SIGTERM after 1s": {
			delayBeforeSendingSignal: time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodContext,
		},
		"context - send SIGINT after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGINT,
			shutdownMethod:           shutdownMethodContext,
		},
		"context - send SIGTERM after 2s": {
			delayBeforeSendingSignal: 2 * time.Second,
			signalToSend:             syscall.SIGTERM,
			shutdownMethod:           shutdownMethodContext,
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			stdout, stderr, elapsed, err := run(os.Args[0], shutdownMethodWait, tt.delayBeforeSendingSignal, tt.signalToSend)
			if err != nil {
				t.Fatalf("unexpected failure: %s\nstdout:\n%s\nstderr:\n%s", err, stdout, stderr)
			}
			res := &result{}
			if err := res.decode([]byte(stdout)); err != nil {
				t.Fatalf("failed to decode the results from stdout: %s\nstdout:\n%s", err, stdout)
			}
			if wantMethod, gotMethod := shutdownMethodWait, res.executedMethod; wantMethod != gotMethod {
				t.Fatalf("expected to have method %q but got %q", wantMethod, gotMethod)
			}
			if elapsed < tt.delayBeforeSendingSignal {
				t.Fatalf("time took to run the shutdown method is less than expected. expected: %s, got: %s", tt.delayBeforeSendingSignal, elapsed)
			}
			inProcessElapsed := res.stoppedAt.Sub(res.startedAt)
			t.Logf("executing and stopping the process took %s and the logic inside the process ran for %s", elapsed, inProcessElapsed)
		})
	}
}

func run(cmdPath string, method string, signalAfter time.Duration, signal os.Signal) (string, string, time.Duration, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Command(cmdPath)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = []string{fmt.Sprintf("%s=%s", envKeyForShutdown, method)}
	startedAt := time.Now()
	if err := cmd.Start(); err != nil {
		return "", "", -1, err
	}

	var waitErr error
	cmdDoneCh := make(chan struct{})
	go func() {
		waitErr = cmd.Wait()
		close(cmdDoneCh)
	}()

	if signalAfter != 0 {
		select {
		case <-time.After(signalAfter):
			if err := cmd.Process.Signal(signal); err != nil {
				return "", "", -1, err
			}
		case <-cmdDoneCh:
		}
	}
	<-cmdDoneCh

	return stdout.String(), stderr.String(), time.Since(startedAt), waitErr
}

type result struct {
	startedAt      time.Time
	stoppedAt      time.Time
	executedMethod string
}

func (r *result) encode() string {
	var b bytes.Buffer
	b.WriteString(r.executedMethod)
	b.WriteString("\n")
	b.WriteString(r.startedAt.Format(time.RFC3339Nano))
	b.WriteString("\n")
	b.WriteString(r.stoppedAt.Format(time.RFC3339Nano))
	return b.String()
}

func (r *result) decode(in []byte) error {
	buf := bufio.NewScanner(bytes.NewReader(in))
	var idx int
	for buf.Scan() {
		t := buf.Text()
		idx++
		switch idx {
		case 1:
			r.executedMethod = t
		case 2:
			tm, err := time.Parse(time.RFC3339Nano, t)
			if err != nil {
				return fmt.Errorf("could not decode start time from the result: %w", err)
			}
			r.startedAt = tm
		case 3:
			tm, err := time.Parse(time.RFC3339Nano, t)
			if err != nil {
				return fmt.Errorf("could not decode stop time from the result: %w", err)
			}
			r.stoppedAt = tm
		default:
			return fmt.Errorf("result can decode only 3 lines of data")
		}
	}
	if idx != 3 {
		return fmt.Errorf("expected to read 3 lines of data but got only %d", idx)
	}
	return nil
}
