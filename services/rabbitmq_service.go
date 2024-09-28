package services

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type Message struct {
	Body string
}

type RabbitMQService interface {
	Publish(queueName string, message Message) error
	Consume(queueName string, handler func(message Message) error) error
	Close() error
}

type rabbitMQService struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQService(amqpURL string) (RabbitMQService, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close() // Close the connection if channel creation fails
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &rabbitMQService{
		connection: conn,
		channel:    ch,
	}, nil
}

// Publish a message to a queue
func (r *rabbitMQService) Publish(queueName string, message Message) error {

	// Declare the queue before publishing (durable, not auto-deleted)
	_, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Publish the message
	err = r.channel.Publish(
		"",        // exchange
		queueName, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message.Body),
		})
	if err != nil {
		return err
	}

	log.Println("Message published successfully")
	return nil
}

// Consume messages from a queue
func (r *rabbitMQService) Consume(queueName string, handler func(message Message) error) error {

	// Consume messages
	msgs, err := r.channel.Consume(
		queueName,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	// Process messages in a goroutine
	go func() {
		for msg := range msgs {
			err := handler(Message{Body: string(msg.Body)})
			if err != nil {
				log.Println("Error handling message:", err)
			}
		}
	}()

	return nil
}

// Close the connection and channel
func (r *rabbitMQService) Close() error {
	if err := r.channel.Close(); err != nil {
		return err
	}
	return r.connection.Close()
}
