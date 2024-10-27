# Timestone 🗿

Timestone is a library to create deterministic and easy-to-understand unit tests for time-dependent, concurrent go 
code. Existing libraries such as [Quartz](https://github.com/coder/quartz) or [Clock](https://github.com/benbjohnson/clock) show the need for such a tool, yet have various 
shortcomings, for example not being able to reliably prevent race-conditions in tests, or being difficult to read and 
understand when used.

### Goals

This library is built around the following primary design goals:

- 🤌 Eliminate flaky unit tests
- 🧹Keep unit tests free of boilerplate syncing code and focussed on assumptions and expectations
- 🐭 As little invasive as possible to the tested implementation

Secondary goals are a good separation of concerns and extensive test coverage for the library itself.

### Design principles

To offer a high-level API that keeps manual syncing code out of unit tests, Timestone takes an opinionated, 
use-case-oriented approach rather than attempting to substitute low-level runtime primitives like timers and tickers. 
However this approach can be limiting, and there may be a need to extend the library’s public interface to support 
additional use cases. For instance, Timestone’s public model already includes passing the commonly used 
`context.Context` but the `simulation` implementation doesn't always respect it. Another example is a cron syntax for 
scheduling recurring tasks which could be easily integrated but is not included yet.

### Limitations

At its current stage, Timestone is fully functional and supports all possible use cases under the assumption that the
computing time for scheduled go routines is not a concern. As a consequence, when testing the time inside actions is
always fixed to an instant and won't pass or change for the duration of the action's execution.

Another major limitation is the lack to fully support contexts: Contexts with a deadline from the `context` standard
libraray package are currently not supported for technical reasons.

Both issues are linked and will be considered in upcoming releases.

## Concepts

To achieve its goals, Timestone aims to encapsulate concurrency. Instead of directly invoking goroutines, the library 
provides a `Scheduler` interface with methods for scheduling `Action`s, such as one-time or recurring tasks. There are 
two implementations of the `Scheduler`: `system.Scheduler` and `simulation.Scheduler`. While the former uses standard 
library runtime primitives to dispatch actions, the latter employs a run loop to control where actions are scheduled. 
Through various configuration options, the scheduling mode and order of actions can be controlled, and action 
dependencies can be setup.

To see how this works in practice, take a look at the `examples` package, which contains functional test cases that 
serve as integration tests for the Timestone library, as well as demonstrative use cases.

The following sections provide a more detailed explanation:

### Scheduler

One of the main challenges in eliminating race conditions from unit tests is handling goroutines. Non-deterministic by 
nature, goroutines provide no guarantee on the order in which concurrent code will be executed by the Go runtime. To 
address this problem in unit tests, Timestone offers a `Scheduler` interface designed to run code concurrently while 
encapsulating the underlying complexity:

```go
type Scheduler interface {
    Clock
    PerformNow(ctx context.Context, action Action, tags ...string)
    PerformAfter(ctx context.Context, action Action, duration time.Duration, tags ...string)
    PerformRepeatedly(ctx context.Context, action Action, until *time.Time, interval time.Duration, tags ...string)
}
```

Where you would normally call `go func() {...}()`, when working with Timestone you instead use the `PerformNow` method 
of the `Scheduler`. The `PerformAfter` and `PerformRepeatedly` methods offer convenient alternatives to using 
`time.Timer` and `time.Ticker` within goroutines for scheduling function execution.

While the `system.Scheduler` implementation of the `Scheduler` interface uses the mentioned runtime scheduling 
primitives, the `simulation.Scheduler` implementation is where the real magic happens.

Rather than immediately running an action within a goroutine, the `simulation.Scheduler` creates an event generator for 
it. The events it materializes will then be executed from either the `ForwardOne` or `Forward` methods, advancing the 
`simulation.Scheduler`’s clock either to the next event or through all events scheduled to occur within a specified 
`time.Duration`. Additional configuration can be provided for individual actions or entire groups, allowing control over 
the execution order of simultaneous events or injecting dependencies between actions, delaying the execution of certain 
actions until their dependencies have completed.

To provide this level of control, the `simulation.Scheduler` uses a run loop that iterates over all events in a 
well-defined order until no event remains. For each event, its configuration and default settings are considered to 
determine whether it should execute sequentially or asynchronously, if it must wait on other events, or if it will 
register a new event generator the run loop has to wait for.

### Action

An `Action` defines an interface for a function to be executed.

```golang
type Action interface {
    Perform(context.Context)
}
```

The `context.Context` provided to the `Perform` method offers contextual information, that is currently a clock. You can 
either use the included `SimpleAction` as a convenient wrapper or create your own implementation.

### Events and event generators

An `Event` is an internal concept of the `simulation.Scheduler` that combines an `Action` with some identifying `Tags` 
and a `time.Time` that determines when it should be executed. These events are produced from actions by 
`simulation.EventGenerator`s. For example, when calling `simulation.Scheduler.PerformRepeatedly`, a corresponding event 
generator is registered, which repeatedly materializes events into the event queue according to its settings.

When using the `simulation.Scheduler` for deterministic unit tests, you configure events by providing 
`EventConfiguration`s. These configurations can target events by the tags, or more specifically by including their 
 execution time.

### Event generators and event queue

`simulation.EventGenerator`s hold information about at the next and potentially following events materialized by them. 
They are either created when calling one of the `simulation.Scheduler.Perform...` methods or simply by adding 
your own generator implementation to a `simulation.Scheduler` (e.g. if you want to use Timestone for pure simulation 
purposes).

```golang
type EventGenerator interface {
    Pop() *Event
    Peek() Event
    Finished() bool
}
```

This interface is then used by the event queue to materialize and sort new events as they are needed in a stream like
fashion.

Knowing this concept is important when it comes to designing tests for business logic where actions will recursively 
schedule more actions (which might schedule more actions). Imagine you have an action `firstAction` that you want to 
execute asynchronously, which is supposed to schedule a `secondAction` via `simulation.Scheduler.PerformNow`. 
In this case the `simulation.Scheduler`'s run loop needs to wait for the generator later materializing 
`secondAction` to be added – otherwise the run loop might terminate in the next iteration, not knowing yet that a new 
generator will provide another event shortly.

To avoid this race condition, you add a new `config.Config` targeting the `firstAction` that looks probably 
like:

```golang
scheduler.ConfigureEvents(
    config.Config{
        Tags: []string{"firstAction"},
        Adds: []*config.Generator{
            Tags: []string{"secondAction"},
            Count: 1,
        },
    },
)
```

Now after executing every `firstAction` event, the scheduler will pause its run loop until a generator producing 
`secondAction` events has been registered.

## Contributing

This project is still under development, and contributions are welcome. Feel free to fork the repository and submit a PR. 
When submitting a PR, it would be helpful to reference an open issue for better documentation.

Currently, the most important features on the agenda for this project are:
- Pipeline for linting and automatic unit tests before merging
- Support for canceled contexts

## Reporting a Bug

To report a bug, please create an issue ticket. Include sufficient code samples and contextual information to reproduce 
the bug. If you can provide a fix, it will be greatly appreciated.



