# Logging

This application implements structured logging using the [logr interface] along
with the [zap logger].

There are many logging solutions in Go, including the logging package in the
standard library. In particular, logr is very popular in the Kubernetes
ecosystem with a well-thought-out interface. Many projects implement their own
logging interfaces to avoid vendor lock, but it is hard to come up with a good
design.

## Configuration

There are two configuration attributes that affect the application logger:

```toml
# When enabled, logs with V-levels greater than zero are shown (debug-y logs).
# When disabled, only logs with V-level 0 are shown.
# Note: v-level verbosity cannot be arbitrarily chosen, but can supported later.
verbose = true

# Use the color-enabled logging encoder.
# Do not use in production, it disables JSON-encoded structured logging.
debug = true
```

## Usage

See [typical usage] for more details. The logr docs are great, read them!

Example:

```go
// Log with level 0, i.e. "you always want to see this".
logger.Info("Hi there!", "time", time.Now())

// Log with level 1, i.e. "you might possibly want to turn off".
logger.V(1).Info("Hello.")

// Log error.
logger.Error(err, "It was not possible to open the file", "path", path)
```

## Logging in the Temporal SDK

There are three levels of logging in the Temporal SDK that we care about.

### Logging from Temporal client

The Temporal client writes its own logs and by default it uses an internal
logger that writes to the standard error stream. We've injected our own logger
with an adapter, e.g.:

```go
c, err := client.Dial(client.Options{
	HostPort:  cfg.Temporal.Address,
	Namespace: cfg.Temporal.Namespace,
	Logger:    temporal.Logger(logger.WithName("temporal")),
})
```

Because Temporal uses semantic levels of logging, we are translating their
errors to logr's errors, `Info` and `Warn` to `V(0)`, `Debug` to `V(1)`.

### Logging from Temporal workflow functions

Workflow code is special because it can be replayed by Temporal from the event
history. This may happen if a workflow worker crashes and Temporal is trying to
resume operations.

When logging, it's best to use the workflow replay-aware [Temporal Go SDK
logger] which uses semantic levels of logging, e.g. `Info` or `Debug`. Other
than that, the API is not too different from logr:

```go
func LoggingWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Hello!", "key", "value")
	logger.Debug("Debugging...", "key", "value")
	if err := fn(); err != nil {
		logger.Error("Error running function.", "err", err)
	}
	return nil
}
```

It is possible to inject our logger using a workflow interceptor, but some
additional code would be required to make it aware of workflow replays, which
is a detail made available to workflow code via the `workflow.Context`.

### Logging from Temporal activity functions

Using an interceptor, we inject the application logger (`logr.Logger`) via the
activity context. Use as follows:

```go
func LoggingActivity(ctx context.Context) error {
	logger := temporal.GetLogger(ctx)
	logger.Info("Hello world!", "time", time.Now())
	logger.V(1).Info("Debugging...")
	if err := fn(); err != nil {
		logger.Error(err, "Error running function.")
	}
	return nil
}
```

It is also possible to inject the logger by other means, e.g. activity structs
can have a `logger logr.Logger` field that is shared with the execute method.

Don't use the [Temporal Go SDK logger] if possible.

## What's the right amount of logging?

We want to avoid excessive error logging. As [Dave Cheney suggested], there are
just two things that we should log:

- _Things that users care about when using your software._<br />
  Mostly errors, using `logger.Error()`. Use `logger.Info()` very sparely.
- _Things that developers care about when they are developing or debugging
  software._<br/>
  Use V-Levels (gte 1), e.g.: `logger.V(1).Info()`.

## How do I choose my V-levels? (from the logr docs)

This is basically the only hard constraint: increase V-levels to denote more
verbose or more debug-y logs.

Otherwise, you can start out with 0 as "you always want to see this", 1 as
"common logging that you might possibly want to turn off", and 10 as "I would
like to performance-test your log collection stack."

Then gradually choose levels in between as you need them, working your way down
from 10 (for debug and trace style logs) and up from 1 (for chattier info-type
logs).

[logr interface]: https://github.com/go-logr/logr
[zap logger]: https://github.com/uber-go/zap
[typical usage]: https://github.com/go-logr/logr#typical-usage
[temporal go sdk logger]: https://github.com/temporalio/sdk-go/blob/HEAD/log/logger.go
[dave cheney suggested]: https://dave.cheney.net/2015/11/05/lets-talk-about-logging
