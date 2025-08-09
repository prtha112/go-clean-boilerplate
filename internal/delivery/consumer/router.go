package consumer

import (
	"context"
	"log"
	"sync"

	"go-clean-boilerplate/internal/domain"
	"go-clean-boilerplate/pkg/kafka"
)

type worker struct {
	consumer kafka.KafkaConsumer
	handler  domain.Handler
	name     string
}

// Router centralizes Kafka consumers and their handlers, mirroring the HTTP router pattern.
type Router struct {
	workers []worker
}

func NewRouter() *Router {
	return &Router{workers: make([]worker, 0)}
}

// Register adds a consumer-Handler pair to the router.
func (r *Router) Register(consumer kafka.KafkaConsumer, handler domain.Handler, name string) {
	r.workers = append(r.workers, worker{consumer: consumer, handler: handler, name: name})
}

// Start begins consuming Kafka messages for all registered workers and dispatching them to handlers.
func (r *Router) Start(ctx context.Context) {
	if len(r.workers) == 0 {
		log.Println("No Kafka workers registered; nothing to start")
		return
	}

	log.Printf("Starting %d Kafka worker(s)...", len(r.workers))

	var wg sync.WaitGroup
	wg.Add(len(r.workers))

	for idx := range r.workers {
		w := r.workers[idx]
		go func(w worker) {
			defer wg.Done()
			for {
				msg, err := w.consumer.ReadMessage()
				if err != nil {
					log.Printf("[%s] Kafka read error: %v", w.name, err)
					continue
				}

				if err := w.handler.Handle(ctx, msg.Key, msg.Value); err != nil {
					log.Printf("[%s] Handler error: %v", w.name, err)
					continue
				}
			}
		}(w)
	}

	// Block until all workers exit (infinite loop unless context-driven exit is implemented)
	wg.Wait()
}

// Close releases any resources used by the router.
func (r *Router) Close() error {
	var firstErr error
	for _, w := range r.workers {
		if err := w.consumer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
