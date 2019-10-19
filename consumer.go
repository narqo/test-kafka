package main

import (
	"context"
	"log"
	"sync"

	"github.com/Shopify/sarama"
)

type MessageHandler interface {
	Handle(ctx context.Context, msg *sarama.ConsumerMessage) error
}

type basicHandler struct{}

func (h *basicHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage) error {
	log.Printf("MessageHandler: got message: topic %q, value %s", msg.Topic, msg.Value)
	return nil
}

type KafkaConsumer struct {
	ready chan struct{}

	msgs chan message
	wg   sync.WaitGroup
}

func NewKafkaConsumer(size int, h MessageHandler) *KafkaConsumer {
	c := &KafkaConsumer{
		ready: make(chan struct{}),
		msgs:  make(chan message, size),
	}

	for i := 0; i < cap(c.msgs); i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()

			c.consume(h)
		}()
	}

	return c
}

type message struct {
	ctx context.Context
	msg *sarama.ConsumerMessage
}

func (c *KafkaConsumer) consume(h MessageHandler) {
	for msg := range c.msgs {
		// TODO(narqo): report handling error back to ConsumerGroupHandler
		h.Handle(msg.ctx, msg.msg)
	}
}

func (c *KafkaConsumer) Setup(ses sarama.ConsumerGroupSession) error {
	log.Printf("consumer: session setup: member-id %q, claims %v", ses.MemberID(), ses.Claims())
	close(c.ready)
	return nil
}

func (c *KafkaConsumer) Cleanup(ses sarama.ConsumerGroupSession) error {
	log.Printf("consumer: session cleanup: member-id %q, claims %v", ses.MemberID(), ses.Claims())
	c.ready = make(chan struct{})
	return nil
}

func (c *KafkaConsumer) ConsumeClaim(ses sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("consumer: message claimed: topic %q, value %s, ts %s", msg.Topic, msg.Value, msg.Timestamp)
		c.msgs <- message{ses.Context(), msg}
		ses.MarkMessage(msg, "")
	}
	return nil
}

func (c *KafkaConsumer) Close() {
	log.Printf("consumer: closing")
	select {
	case <-c.ready:
		// close is called after the last ConsumeClaim call,
		// so it's safe to close msgs here
		close(c.msgs)
	default:
	}
}
