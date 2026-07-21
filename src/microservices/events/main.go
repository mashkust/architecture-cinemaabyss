package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

type EventResponse struct {
	Status    string `json:"status"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Event     Event  `json:"event"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		broker = "kafka:9092"
	}

	go startConsumer(broker, "movie-events", "events-movie-group")
	go startConsumer(broker, "user-events", "events-user-group")
	go startConsumer(broker, "payment-events", "events-payment-group")

	http.HandleFunc("/api/events/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":true}`))
	})

	http.HandleFunc("/api/events/movie", createEventHandler(broker, "movie-events", "movie"))
	http.HandleFunc("/api/events/user", createEventHandler(broker, "user-events", "user"))
	http.HandleFunc("/api/events/payment", createEventHandler(broker, "payment-events", "payment"))

	log.Printf("events-service started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func createEventHandler(broker, topic, eventType string) http.HandlerFunc {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"method not allowed"}`))
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid request body"}`))
			return
		}

		eventID := eventType + "-" + time.Now().Format("20060102150405.000000000")
		event := Event{
			ID:        eventID,
			Type:      eventType,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Payload:   payload,
		}

		data, err := json.Marshal(event)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"failed to marshal event"}`))
			return
		}

		msg := kafka.Message{
			Key:   []byte(event.ID),
			Value: data,
		}

		err = writer.WriteMessages(context.Background(), msg)
		if err != nil {
			log.Printf("failed to publish event to %s: %v", topic, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"failed to publish event"}`))
			return
		}

		log.Printf("[producer] topic=%s event=%s", topic, string(data))

		resp := EventResponse{
			Status:    "success",
			Partition: 0,
			Offset:    0,
			Event:     event,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

func startConsumer(broker, topic, groupID string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	log.Printf("[consumer] started for topic=%s", topic)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("[consumer] read error topic=%s err=%v", topic, err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("[consumer] topic=%s partition=%d offset=%d value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
	}
}
