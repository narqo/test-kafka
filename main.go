package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
)

const (
	groupID = "go-consumer-group0"
	topic   = "go-topic0"

	numConsumers = 5
)

type config struct {
	kafkaVersion string
	kafkaBrokers []string
}

func main() {
	var conf config

	flag.StringVar(&conf.kafkaVersion, "kafka.version", "1.1.0", "kafka protocol version")
	kb := flag.String("kafka.brokers", "localhost:9092", "kafka bootstrap brokers")

	flag.Parse()

	if *kb == "" {
		log.Fatal("no kafka bootstrap broker")
	}
	conf.kafkaBrokers = strings.Split(*kb, ",")

	sarama.Logger = log.New(os.Stdout, "sarama ", log.LstdFlags)

	if err := run(conf); err != nil {
		log.Fatal(err)
	}
}

func run(conf config) error {
	version, err := sarama.ParseKafkaVersion(conf.kafkaVersion)
	if err != nil {
		return fmt.Errorf("could not parse kafka version %q: %w", conf.kafkaVersion, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	kafkaConf := sarama.NewConfig()
	kafkaConf.Version = version

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	errs := make(chan error, 1)

	group, err := sarama.NewConsumerGroup(conf.kafkaBrokers, groupID, kafkaConf)
	if err != nil {
		return fmt.Errorf("could not create consumer group %q, brokers %v: %w", groupID, conf.kafkaBrokers, err)
	}
	defer group.Close()

	go func() {
		c := NewKafkaConsumer(numConsumers, &basicHandler{})
		defer c.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := group.Consume(ctx, []string{topic}, c); err != nil {
				errs <- err
				return
			}
		}
	}()

	select {
	case <-sigs:
		cancel()
		return nil
	case err := <-errs:
		return err
	}
}
