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

type item[E any] struct {
	Data E
	Err  error
}

type Channel[E any] interface {
	Send(E) bool
	SendErr(error) bool
}

type channelIter[E any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	c      chan item[E]
	item   item[E]
}

func (ci *channelIter[E]) Send(e E) bool {
	select {
	case ci.c <- item[E]{Data: e}:
		return false
	case <-ci.ctx.Done():
		return true
	}
}

func (ci *channelIter[E]) SendErr(e error) bool {
	select {
	case ci.c <- item[E]{Err: e}:
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
		c:      make(chan item[E]),
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

type Result[E any] struct {
	Value E
	Err   error
}

func SendToChannel[E any](iter Iterator[E], c chan Result[E]) {
	for iter.Next() {
		c <- Result[E]{
			Value: iter.Get(),
		}
	}
	if err := iter.Err(); err != nil {
		c <- Result[E]{
			Err: err,
		}
	}
	close(c)
}
