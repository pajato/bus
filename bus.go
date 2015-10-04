// Copyright 2015 Pajato Group Inc. All rights reserved.  Use of this
// source code is governed by the GNU GPL v3 license.  See the LICENSE
// file.

// Package bus is a subscription based (Publisher/Subscriber pattern)
// communication subsystem based loosely on the Java/Android Otto
// library, which in turn is based on Guava.  An inspirational credit
// must also be given to the mBassador project at
// https://github.com/bennidi/mbassador
//
// A good way to think of this bus is to imagine an infinitely large
// tour bus that picks up anonymous riders carrying a typed collection
// of (possibly empty) data and then transports them to be handled by
// other subsystems (goroutines) or functions executed by the bus
// goroutine.
//
// The typed collection of data is a Payload.
package bus

import (
	"fmt"
	"log"
	"time"
)

// A Payload type will provide a type and some, possibly empty, data.
// It might be most likely to build an event type to carry a payload
// on the bus but is by no means limited to just events.
type Payload interface {
	Type() string
	Data() map[string]interface{}
}

// A Handler instance will be called by the bus when an event of the
// type registered for it is posted to the bus.  Handlers are
// registered via the Subscribe method.
type Handler func(p Payload) error

// The flag type acts as a base type for Bus constants.
type flag int

const (
	synchronous  flag = 1 << iota
	asynchronous flag = 1 << iota
)

// A rider carries a payload and a delivery mode.
type rider struct {
	payload Payload
	mode    flag
}

// A Bus instance will communicate Payload objects to other goroutines
// using a channel and/or a list of handlers.
type Bus struct {
	pubchan  chan rider
	subchans map[string][]chan Payload
	handlers map[string][]Handler
	flags    map[Payload]flag
}

// Log a message using the configuration established by the bus package.
func (b Bus) Log(message string) {
	log.Print(message)
}

// Post will asynchonously notify all subscribers that a payload of a
// certain type is available.
func (b Bus) Post(p Payload) error {
	log.Printf("Posting payload of type: %v.\n", p.Type())
	r := rider{p, asynchronous}
	b.pubchan <- r
	return nil
}

// PostAndWait synchronously notifies all subscribers.
func (b Bus) PostAndWait(p Payload) error {
	log.Printf("Posting payload of type: %v.\n", p.Type())
	r := rider{p, synchronous}
	b.pubchan <- r
	return nil
}

// AddHandlers will register one or more handlers for a given payload
// type.  Registering no handlers is an error.
func (b Bus) AddHandlers(typ string, fns ...Handler) error {
	if len(fns) > 0 {
		b.handlers[typ] = append(b.handlers[typ], fns...)
		return nil
	}
	message := "Argument error: at least one handler must be registered."
	return &busError{time.Now(), message}
}

// AddChannel will register a channel for a given payload type.
func (b Bus) AddChannel(typ string, c chan Payload) {
	if _, ok := b.subchans[typ]; !ok {
		b.subchans[typ] = []chan Payload{c}
	} else {
		subchans := b.subchans[typ]
		b.subchans[typ] = append(subchans, c)
	}
}

// New will create a Bus object with a channel on which to post a
// Payload object and empty sets of subscriber functions and
// subscriber channels.  Lastly, the new Bus object will run a traffic
// cop steering posted payloads to the handlers that will deal with
// them.
func New() Bus {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Creating a new bus that runs a traffic cop to handle posted payloads.")
	b := new(Bus)
	b.pubchan = make(chan rider)
	b.subchans = make(map[string][]chan Payload)
	b.handlers = make(map[string][]Handler)
	go b.run()

	return *b
}

type busError struct {
	When time.Time
	What string
}

// Error provides a hook to access the latest error.
func (e *busError) Error() string {
	return fmt.Sprintf("at %v, %s", e.When, e.What)
}

// Run the bus to listen for posts.
func (b Bus) run() {
	log.Println("Bus is running.")
	for r := range b.pubchan {
		// Distribute the payload carried by the rider to the
		// registered handlers and subscribers.
		log.Printf("Broadcasting payload with type: %v, %v.\n", r.payload.Type(), b.modestring(r.mode))
		if r.mode == asynchronous {
			// Deliver the payload carried by the rider asynchronously.
			go b.deliver(r)
		} else {
			// Deliver the payload carried by the rider synchronously.
			b.deliver(r)
		}
	}
	log.Println("Bus is stopping.")
}

func (b Bus) deliver(r rider) {
	// First deliver the payload to the handlers.
	typ := r.payload.Type()
	for i, h := range b.handlers[typ] {
		log.Printf("Processing payload with type: %v, and handler at index: %v.\n", typ, i)
		err := h(r.payload)
		if err != nil {
			log.Printf("Handler failed: %v.\n", h)
		}
	}
	for i, c := range b.subchans[typ] {
		// Now deliver the payload to the subsystems.
		log.Printf("Processing payload with type: %v, and channel at index: %v.\n", typ, i)
		c <- r.payload
	}
}

func (b Bus) modestring(f flag) string {
	if (f & asynchronous) == asynchronous {
		return "asynchronously"
	}
	return "synchronously"
}
