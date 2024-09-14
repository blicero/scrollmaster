// /home/krylon/go/src/krylib/msgbuf.go
// -*- mode: go; coding: utf-8; -*-
// Created on 13. 09. 2019 by Benjamin Walkenhorst
// (c) 2019 Benjamin Walkenhorst
// Time-stamp: <2022-09-06 19:10:04 krylon>

// TODO Now that Go has Generics, I might want to rewrite this thing to
//      carry a generic payload.

package krylib

import (
	"sync"
	"time"
)

// TimestampFormat is a commonly used format time values
const TimestampFormat = "2006-01-02 15:04:05"

// Message is a ... message that can sent and received via the MessageBuffer.
type Message struct {
	Msg   string
	Stamp time.Time
}

// StampString returns the Message's timestamp as a string.
func (m *Message) StampString() string {
	return m.Stamp.Format(TimestampFormat)
} // func (self *Message) StampString() string

// MessageBuffer is a queue-like structure.
type MessageBuffer struct {
	messages []Message
	lock     sync.Mutex
	queue    chan Message
	empty    *sync.Cond
	running  bool
	glock    sync.Mutex
}

// CreateMessageBuffer creates a MessageBuffer.
func CreateMessageBuffer() *MessageBuffer {
	var mbuf = &MessageBuffer{
		messages: make([]Message, 0),
		queue:    make(chan Message),
		running:  true,
	}

	mbuf.empty = sync.NewCond(&mbuf.lock)

	go mbuf.worker()

	return mbuf
} // func CreateMessageBuffer() *MessageBuffer

// Running returns true if the MessgeBuffer is active.
func (b *MessageBuffer) Running() bool {
	var running bool
	b.glock.Lock()
	running = b.running
	b.glock.Unlock()
	return running
} // func (self *MessageBuffer) Running() bool

// Stop tells the message buffer to cease activity.
func (b *MessageBuffer) Stop() {
	b.glock.Lock()
	b.running = false
	b.glock.Unlock()
} // func (self *MessageBuffer) Stop()

func (b *MessageBuffer) worker() {
	for {
		var stillGoing, sendSignal bool
		var msg Message

		b.glock.Lock()
		stillGoing = b.running
		b.glock.Unlock()

		if !stillGoing {
			return
		}

		msg = <-b.queue

		b.lock.Lock()
		sendSignal = len(b.messages) == 0

		b.messages = append(b.messages, msg)
		b.lock.Unlock()
		if sendSignal {
			b.empty.Signal()
		}
	}
} // func (b *MessageBuffer) worker()

// AddMessage adds a message to the Buffer
func (b *MessageBuffer) AddMessage(msg string) {
	var m = Message{
		Msg:   msg,
		Stamp: time.Now(),
	}

	b.queue <- m
} // func (self *MessageBuffer) AddMessage(msg string)

// Empty returns true if the queue is currently empty.
func (b *MessageBuffer) Empty() bool {
	b.lock.Lock()
	cnt := len(b.messages)
	b.lock.Unlock()
	return cnt == 0
} // func (b *MessageBuffer) Empty() bool

// Count returns the number of messages currently queued in the buffer.
func (b *MessageBuffer) Count() int {
	b.lock.Lock()
	cnt := len(b.messages)
	b.lock.Unlock()
	return cnt
} // func (b *MessageBuffer) Count() int

// GetOneMessage retrieves a single message from the MessageBuffer.
func (b *MessageBuffer) GetOneMessage() *Message {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.messages) == 0 {
		b.empty.Wait()
	}

	var m = new(Message)

	*m = b.messages[0]

	b.messages = b.messages[1:len(b.messages)]

	return m
} // func (b *MessageBuffer) GetOneMessage() *Message

// GetAllMessages retrieves all Messages that are currently in the queue.
func (b *MessageBuffer) GetAllMessages() []Message {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.messages) == 0 {
		//b.empty.Wait()
		return make([]Message, 0)
	}

	//var m []Message = make([]Message, len(b.messages))

	var m = b.messages
	b.messages = make([]Message, 0)

	return m
} // func (b *MessageBuffer) GetAllMessages() []Mesage

// Local Variables:  //
// compile-command: "go generate && go vet && go build -v -p 16 && gometalinter && go test -v" //
// End: //
