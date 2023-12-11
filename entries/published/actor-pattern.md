---
title: actor-pattern
published: 2023-07-31
revision: 2023-07-31
tags: go
excerpt: Actors provide a nice and simple pattern to develop concurrent programs.
---

The actor model helps us deal with concurrent programs in a way that does not
require locks.

In a message-driven approach, _actors_, the main building block of this model,
receive messages and process them sequentially.

This implementation is made quite simple and easy to follow by the fact that
our messages are functions that we send to a channel.

```go
// Package actor provides a simple actor implementation.
package actor

import "fmt"

type action func()

// Actor is an entity that processes actions sent to it via its inbox.
type Actor struct {
	Name      string
	InboxSize int
	inbox     chan action
}

// ConfigFn is the signature of funcs that configure an actor.
type ConfigFn func(a *Actor)

const defaultInboxSize = 16

// New creates a new actor and starts its message processing loop.
func New(fns ...ConfigFn) *Actor {
	a := &Actor{}
	a.InboxSize = defaultInboxSize
	for _, fn := range fns {
		fn(a)
	}
	a.inbox = make(chan action, a.InboxSize)
	go a.start()
	return a
}

// SetInboxSize sets the size of the actor's inbox, which is a chan of funcs.
// The actor's inbox could be an unbuffered channel, but that would block
// clients from sending messages to the actor while it is processing a message.
// By buffering the channel we also ensure that messages up to the size of
// the buffer are processed in the order they are received.
// Hence, the recommendation is to use a buffered channel, but you can use an
// unbuffered one by setting the size to 0.
func SetInboxSize(size int) ConfigFn {
	return func(a *Actor) {
		a.InboxSize = size
	}
}

// SetName provides a name for the actor.
func SetName(name string) ConfigFn {
	return func(a *Actor) {
		a.Name = name
	}
}

// start kicks off the actor's message processing loop.
func (a *Actor) start() {
	for {
		fn := <-a.inbox
		fn()
	}
}

// Do sends an action to the actor's inbox for processing.
func (a *Actor) Do(fn action) {
	a.inbox <- fn
}

// Try sends an action to the actor's inbox for processing, but does not block
// if the inbox is full. It just discards the action.
func (a *Actor) Try(fn action) bool {
	select {
	case a.inbox <- fn:
		return true
	default:
		return false
	}
}

func (a *Actor) String() string {
	return fmt.Sprintf("Actor:%s{InboxSize: %d}", a.Name, a.InboxSize)
}
```

And you would use it like so:

```go
package main

import (
	"actors/actor"
	"fmt"
	"time"
)

func main() {
	// Create actor with the default inbox size
	a := actor.New()

	// Create actor with a custom inbox size and a name
	// a := actor.New(actor.SetInboxSize(4), actor.SetName("MyActor"))

	// Send actions to the actor
	a.Do(func() { fmt.Println("Hello") })
	a.Do(func() { fmt.Println("World") })

	// Send action and wait for a response
	response := make(chan int)
	a.Do(func() {
		ans := 40 + 2
		response <- ans
	})
	fmt.Println(<-response)

	// Send action but discard it if Actor's inbox is full
	sent := a.Try(func() { fmt.Println("Goodbye") })
	fmt.Println("Sent action to actor?", sent)

	fmt.Println(a)

	// Simulate a task that keeps the program alive
	time.Sleep(1 * time.Second)
}
```

Hope you find this useful and play around with the actor model if you haven't
done so yet.
