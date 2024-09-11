package examples

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/metamogul/timestone/simulation"
	"github.com/stretchr/testify/require"
)

const simulateLoadMilliseconds = 100

type processingCache struct {
	content   map[string]string
	contentMu sync.RWMutex
}

func newProcessingCache() *processingCache {
	return &processingCache{
		content: make(map[string]string),
	}
}

func (p *processingCache) set(key string, value string) {
	p.contentMu.Lock()
	defer p.contentMu.Unlock()

	p.content[key] = value
}

func (p *processingCache) get(key string) (value string) {
	p.contentMu.RLock()
	defer p.contentMu.RUnlock()

	return p.content[key]
}

func (p *processingCache) getContent() (content map[string]string) {
	p.contentMu.RLock()
	defer p.contentMu.RUnlock()

	return p.content
}

type fooProcessor struct {
	cache     *processingCache
	scheduler timestone.Scheduler
}

func newFooProcessor(cache *processingCache, scheduler timestone.Scheduler) *fooProcessor {
	return &fooProcessor{
		cache:     cache,
		scheduler: scheduler,
	}
}

func (a *fooProcessor) invoke(context.Context) {
	for key, value := range a.cache.getContent() {
		time.Sleep(time.Duration(rand.Int64N(simulateLoadMilliseconds)) * time.Millisecond)
		a.cache.set(key, value+"foo")
	}
	fmt.Println("leaving add foo")
}

type barProcessor struct {
	cache     *processingCache
	scheduler timestone.Scheduler
}

func newBarProcessor(cache *processingCache, scheduler timestone.Scheduler) *barProcessor {
	return &barProcessor{
		cache:     cache,
		scheduler: scheduler,
	}
}

func (m *barProcessor) invoke(ctx context.Context) {
	wg := &sync.WaitGroup{}

	for key, value := range m.cache.getContent() {
		wg.Add(1)

		time.Sleep(time.Duration(rand.Int64N(simulateLoadMilliseconds)) * time.Millisecond)
		m.cache.set(key, value+"bar")

		m.scheduler.PerformNow(ctx, timestone.NewSimpleAction(
			func(context.Context) {
				time.Sleep(time.Duration(rand.Int64N(simulateLoadMilliseconds)) * time.Millisecond)

				value := m.cache.get(key)
				m.cache.set(key, value+"baz")
				wg.Done()
			},
			"barPostprocessingBaz",
		))
	}
	wg.Wait()
}

type app struct {
	ctx context.Context

	scheduler    timestone.Scheduler
	cache        *processingCache
	fooProcessor *fooProcessor
	barProcessor *barProcessor
}

func newApp(scheduler timestone.Scheduler) *app {
	cache := newProcessingCache()
	fooProcessor := newFooProcessor(cache, scheduler)
	barProcessor := newBarProcessor(cache, scheduler)

	return &app{
		ctx: context.Background(),

		scheduler:    scheduler,
		cache:        cache,
		fooProcessor: fooProcessor,
		barProcessor: barProcessor,
	}
}

func (a *app) seedCache() {
	a.cache.set("bort", "")
	a.cache.set("burf", "")
	a.cache.set("bell", "")
	a.cache.set("bick", "")
	a.cache.set("bams", "")
}

func (a *app) run() {
	a.scheduler.PerformRepeatedly(a.ctx, timestone.NewSimpleAction(a.fooProcessor.invoke, "fooProcessing"), nil, time.Hour)
	a.scheduler.PerformRepeatedly(a.ctx, timestone.NewSimpleAction(a.barProcessor.invoke, "barProcessing"), nil, time.Hour)
}

func TestApp(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	testcases := []struct {
		name               string
		configureScheduler func(s *simulation.Scheduler)
		result             string
	}{
		{
			name: "foo before bar",
			configureScheduler: func(s *simulation.Scheduler) {
				s.SetDefaultMode(simulation.ExecModeAsync)
				s.ConfigureEvent("barProcessing", nil, simulation.EventConfiguration{
					WaitForActions:     []string{"fooProcessing"},
					WantsNewGenerators: map[string]int{"barPostprocessingBaz": 5},
				})
			},
			result: "foobarbaz",
		},
		{
			name: "foo after bar",
			configureScheduler: func(s *simulation.Scheduler) {
				s.SetDefaultMode(simulation.ExecModeAsync)
				s.ConfigureEvent("fooProcessing", nil, simulation.EventConfiguration{
					WaitForActions: []string{"barProcessing", "barPostprocessingBaz"},
				})
				s.ConfigureEvent("barProcessing", nil, simulation.EventConfiguration{
					WantsNewGenerators: map[string]int{"barPostprocessingBaz": 5},
				})
			},
			result: "barbazfoo",
		},
		{
			name: "foo after bar, sequential",
			configureScheduler: func(s *simulation.Scheduler) {
				s.SetDefaultMode(simulation.ExecModeAsync)
				s.ConfigureEvent("fooProcessing", nil, simulation.EventConfiguration{
					ExecMode: simulation.ExecModeSequential,
					Priority: 2,
				})
				s.ConfigureEvent("barProcessing", nil, simulation.EventConfiguration{
					WantsNewGenerators: map[string]int{"barPostprocessingBaz": 5},
					Priority:           1,
				})
			},
			result: "barbazfoo",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			simulationScheduler := simulation.NewScheduler(now)

			tt.configureScheduler(simulationScheduler)

			a := newApp(simulationScheduler)
			a.seedCache()
			a.run()

			simulationScheduler.Forward(1 * time.Hour)

			for _, value := range a.cache.content {
				require.Equal(t, tt.result, value)
			}
		})
	}
}
