package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"log/slog"
	"strings"
)

var (
	brokers = flag.String("brokers", "127.0.0.1:9092", "Kafka brokers to connect to")
	topic   = flag.String("topic", "test-events", "Kafka topic to publish to / receive from")
)

func main() {
	flag.Parse()

	brokerList := strings.Split(*brokers, ",")
	producer, err := newProducer(brokerList)
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()

	consumer, cErr := newConsumer(brokerList)
	if cErr != nil {
		log.Fatalf("Error creating consumer: %v", cErr)
	}
	defer consumer.Close()

	client := KafkaClient{consumer, producer, *topic}
	client.Send("Hi world")
	client.Read()
}

type KafkaClient struct {
	consumer sarama.Consumer
	producer sarama.SyncProducer
	topic    string
}

func newConsumer(brokerList []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return sarama.NewConsumer(brokerList, config)
}

func (c KafkaClient) Send(message string) error {
	msg := &sarama.ProducerMessage{Topic: c.topic, Value: sarama.StringEncoder(message)}
	partition, offset, err := c.producer.SendMessage(msg)
	if err != nil {
		return err
	}
	slog.Debug("Message sent", "partition", partition, "offset", offset)
	return nil
}

func newProducer(brokerList []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 10
	return sarama.NewSyncProducer(brokerList, config)
}

func (c KafkaClient) Read() error {
	partitions, pErr := c.consumer.Partitions(*topic)
	if pErr != nil {
		return errors.New(fmt.Sprintf("Error getting partitions: %v", pErr))
	}
	slog.Debug("Current high water marks:", c.consumer.HighWaterMarks())
	for _, p := range partitions {
		slog.Debug("Partition: %d\n", p)
		pCons, cErr := c.consumer.ConsumePartition(*topic, p, sarama.OffsetOldest)
		if cErr != nil {
			return errors.New(fmt.Sprintf("Error consuming partition %d: %v", p, cErr))
		}
		for m := range pCons.Messages() {
			strMessage := string(m.Value)
			slog.Info("Received message", "msg", strMessage, "offset", m.Offset)
		}
	}
	return nil
}
