package main

import (
	"flag"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"strings"
)

var (
	brokers = flag.String("brokers", "127.0.0.1:9092", "Kafka brokers to connect to")
	topic   = flag.String("topic", "test-events", "Kafka topic to publish to")
)

func main() {
	flag.Parse()

	brokerList := strings.Split(*brokers, ",")
	producer, err := newProducer(brokerList)
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}
	defer producer.Close()
	Send(producer, *topic)
	consumer, err := newConsumer(brokerList)
	partitions, pErr := consumer.Partitions(*topic)
	defer consumer.Close()
	if pErr != nil {
		log.Fatalf("Error getting partitions: %v", pErr)
	}
	fmt.Println("Current high water marks:", consumer.HighWaterMarks())
	//wg := &sync.WaitGroup{}
	for _, p := range partitions {
		fmt.Printf("Partition: %d\n", p)
		//wg.Add(1)
		pCons, cErr := consumer.ConsumePartition(*topic, p, sarama.OffsetOldest)
		if cErr != nil {
			log.Fatalf("Error consuming partition %d: %v", p, cErr)
		}
		//go func() {
		//	defer wg.Done()
		for m := range pCons.Messages() {
			strMessage := string(m.Value)
			//fmt.Println(strMessage)
			fmt.Printf("Read message from partition %s for offset %v, block timestamp %v, headers %v\n", strMessage, m.Offset, m.Timestamp, m.Headers)
		}
		//}()

	}

	//wg.Wait()

}

func newConsumer(brokerList []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	return sarama.NewConsumer(brokerList, nil)
}

func Send(producer sarama.SyncProducer, topic string) {
	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder("hello again world")})
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
	log.Printf("Message Sent to partition %d with offset %d", partition, offset)
}

func newProducer(brokerList []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.DefaultVersion
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 10
	return sarama.NewSyncProducer(brokerList, config)
}
