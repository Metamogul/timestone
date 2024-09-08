# Timestone 🗿

Timestone is a library to improve unit tests for time-dependent, concurrent go code. Existing libraries such as [Quartz](https://github.com/coder/quartz) or [Clock](github.com/benbjohnson/clock) show the need for such a tool, yet have various shortcomings, for example not being able to reliably prevent race-conditions in tests, or being difficult to read and understand when used.

### Goals

This library is built around the following primary design goals:

- 🤌 Eliminate flaky unit tests
- 🧹Keep unit tests free of boilerplate syncing code and focussed on assumptions and expectations
- 🐭 As little invasive as possible to the tested implementation

Secondary goals are a good separation of concerns and extensive test coverage for the library itself.

### Design principles

To offer a high-level API that keeps manual syncing code out of unit tests, Timestone takes an opinionated, use-case-oriented approach rather than attempting to substitute low-level runtime primitives like timers and tickers. However this approach can be limiting, and there may be a need to extend the library’s public interface to support additional use cases. For instance, Timestone’s public model already includes passing the commonly used `context.Context`; however the `simulation` implementation doesn't alawys respect it. Additionally, a cron syntax for scheduling recurring tasks could be easily integrated but is currently not included.

## Concepts

To achieve its goals, Timestone aims to mock and model concurrency. Instead of directly invoking goroutines, the library provides a `Scheduler` interface with methods for scheduling `Action`s, such as one-time or recurring tasks. There are two implementations of the `Scheduler`: `system.Scheduler` and `simulation.Scheduler`. While the former uses standard library runtime primitives to dispatch actions, the latter employs a run loop to control the execution timing of tasks. Through various configuration options, the scheduling mode and order of tasks can be controlled, and task dependencies can be injected.

To see how this works in practice, take a look at the `examples` package, which contains functional test cases that serve as integration tests for the Timestone library, as well as demonstrative use cases.

The following sections provide a more detailed explanation:

### Scheduler

One of the main challenges in eliminating race conditions from unit tests is handling goroutines. Non-deterministic by nature, goroutines provide no guarantee on the order in which concurrent code will be executed by the Go runtime. To address this problem in unit tests, Timestone offers a `Scheduler` interface designed to run code concurrently while encapsulating the underlying complexity:

```go
type Scheduler interface {
    Clock
    PerformNow(ctx context.Context, action Action)
    PerformAfter(ctx context.Context, action Action, duration time.Duration)
    PerformRepeatedly(ctx context.Context, action Action, until *time.Time, interval time.Duration)
}
```

Instead of using `go func() {...}()`, the `Scheduler` provides the `PerformNow` method. The `PerformAfter` and `PerformRepeatedly` methods offer convenient alternatives to using `time.Timer` and `time.Ticker` within goroutines for scheduling function execution.

While the `system.Scheduler` implementation of the `Scheduler` interface uses the mentioned runtime scheduling primitives, the `simulation.Scheduler` implementation is where the real magic happens.

Rather than immediately running an action within a goroutine, the `simulation.Scheduler` creates an event generator for it. These events can then be scheduled using either the `ForwardOne` or `Forward` methods, advancing the `simulation.Scheduler`’s clock either to the next event or through all events scheduled to occur within a specified `time.Duration`. Additional configuration can be provided for individual actions or entire groups, allowing control over the scheduling order of simultaneous events or injecting dependencies between actions, delaying the execution of certain actions until their dependencies have completed.

To provide this level of control, the `simulation.Scheduler` uses a run loop that iterates over all events in a well-defined order until no events remain. For each event, its configuration and default settings are considered to determine whether it should execute sequentially or asynchronously, or if it must wait on other events.

### Action

An `Action` provides an interface a function to be called along with a name used to link it to a configuration. 

```golang
type Action interface {
	Perform(ActionContext)
	Name() string
}
```

The ActionContext passed to it provides contextual information such as a clock, as well as a control mechanism that facilitates the recursive scheduling of more actions inside an action when using the `simulation.Scheduler`. You can either use the included `SimpleAction` to wrap your code or implement your own.

### Event

An `Event` is a concept internal to the `simulation.Scheduler` and basically fuses an `Action` with a `time.Time` when it's due for execution. When e.g. calling `simulation.Scheduler.PerformRepeatedly`, a matching event generator will be registered, dropping events into the event loop.

When using the `simulation.Scheduler` to build deterministic unit tests, you will configure events by providing `EventConfiguration`s for them, either targetting events only by the name of their embedded action, or also by their execution time.

## Contributing

This project is still in an early stage and contributions are welcome. Feel free to contribute create a fork and issue a PR. When issuing a PR, it would be nice if you could relate it to an open ticket so there's documentation later on.

If you want to contribute, these are the most important features on the current agenda for this project:
- Pipeline for linting and automatic unit tests before merging
- Support for cancelled contexts

## Reporting a bug

To report a bug, please create an issue ticket for it. In the ticket please provide sufficient code samples along with contextual information that makes the bug reproducable. If you can provide a fix, that certainly is welcome.




