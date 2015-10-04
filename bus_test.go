// Copyright 2015 Pajato Group Inc. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

// Package bus is a subscription based (Publisher/Subscriber pattern)
// communication subsystem based loosely on the Java/Android Otto
// library, which in turn is based on Guava.
package bus

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/pajato/event"
)

func TestZeroValues(t *testing.T) {
	b := New()
	log.Print("message")
	log.Print("another message")
	if b.pubchan == nil {
		t.Error("The posting channel did not get created.")
	}
	if b.subchans == nil {
		t.Error("The map of subscription channels did not get created.")
	}
	if b.handlers == nil {
		t.Error("The map of handlers did not get created.")
	}
}

func TestEmptyMaps(t *testing.T) {
	b := New()
	if n := len(b.subchans); n != 0 {
		message := fmt.Sprintf("The map of subscription channels has a non-zero length, %v.", n)
		t.Error(message)
	}
	if n := len(b.handlers); n != 0 {
		message := fmt.Sprintf("The map of handlers has a non-zero length, %v.", n)
		t.Error(message)
	}
}

func TestError(t *testing.T) {
	b := New()
	err := b.AddHandlers("testName")
	if err == nil {
		t.Error("AddChannels did not return an error as expected.")
	} else {
		message := "Argument error: at least one handler must be registered."
		if !strings.HasSuffix(err.Error(), message) {
			t.Error("AddChannels has an invalid eror message:" + message)
		}
	}
}

func TestAddHandlers(t *testing.T) {
	b := New()
	name := "testName"
	b.AddHandlers(name, h1)
	if n := len(b.handlers[name]); n != 1 {
		t.Errorf("The map of handlers should be 1, but is: %v.", n)
	}
	b.AddHandlers(name, h2, h3, h4)
	if n := len(b.handlers[name]); n != 4 {
		t.Errorf("The map of handlers should be 4, but is: %v.", n)
	}
}

func TestRunHandlersWithNoData(t *testing.T) {
	b := New()
	name := "testEvent"
	b.AddHandlers(name, h1, h2, h3, h4)
	e := event.New(name)
	b.Post(e)
}

func TestRunAsynchHandlersWithData(t *testing.T) {
	b := New()
	name := "testEventAsync"
	b.AddHandlers(name, handler)
	e := event.New(name)
	e.Data()["count"] = 23
	if err := b.Post(e); err != nil {
		t.Errorf("The post failed with message: %v.\n", err)
	}
}

func TestRunSynchHandlersWithData(t *testing.T) {
	b := New()
	name := "testEventSync"
	b.AddHandlers(name, handler)
	e := event.New(name)
	e.Data()["count"] = 23
	if err := b.PostAndWait(e); err != nil {
		t.Errorf("The post failed with message: %v.\n", err)
	}
}

func h1(p Payload) error { return nil }
func h2(p Payload) error { return nil }
func h3(p Payload) error { return nil }
func h4(p Payload) error { return nil }

func handler(p Payload) error {
	fmt.Printf("Payload data is: %v.\n", p.Data()["count"])
	return nil
}
