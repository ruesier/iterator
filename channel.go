package iterator

import (
	"context"
	"sync"
)

type Empty[E any] struct{}

func (e Empty[E]) Next() bool { return false }
func (e Empty[E]) Get() E {
	var zero E
	return zero
}
func (e Empty[E]) Err() error { return nil }

type Item[E any] struct {
	Data E
	Err  error
}

type Channel[E any] interface {
	Send(Item[E]) bool
}

type channelIter[E any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	c      chan Item[E]
	item   Item[E]
}

func (ci *channelIter[E]) Send(r Item[E]) bool {
	select {
	case ci.c <- r:
		return false
	case <-ci.ctx.Done():
		return true
	}
}

func (ci *channelIter[E]) Next() bool {
	if ci.item.Err != nil {
		return false
	}
	select {
	case item, open := <-ci.c:
		if !open {
			return false
		}
		ci.item = item
		if ci.item.Err != nil {
			ci.cancel()
		}
		return ci.item.Err == nil
	case <-ci.ctx.Done():
		ci.item.Err = ci.ctx.Err()
		return false
	}
}

func (ci *channelIter[E]) Get() E {
	return ci.item.Data
}

func (ci *channelIter[E]) Err() error {
	return ci.item.Err
}

func NewChannelIterator[E any](_ctx context.Context, generators ...func(Channel[E])) Iterator[E] {
	if len(generators) == 0 {
		return Empty[E]{}
	}
	ctx, cancel := context.WithCancel(_ctx)
	ci := &channelIter[E]{
		ctx:    ctx,
		cancel: cancel,
		c:      make(chan Item[E]),
	}
	if len(generators) == 1 {
		go func() {
			defer close(ci.c)
			generators[0](ci)
		}()
		return ci
	}
	var wg sync.WaitGroup
	wg.Add(len(generators))
	for _, gen := range generators {
		go func(generator func(Channel[E])) {
			defer wg.Done()
			generator(ci)
		}(gen)
	}
	go func() {
		wg.Wait()
		close(ci.c)
	}()
	return ci
}
