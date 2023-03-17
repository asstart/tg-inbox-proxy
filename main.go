package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	tele "gopkg.in/telebot.v3"
)

type opts struct {
	Token                 string `short:"t" long:"token" env:"TG_BOT_TOKEN" description:"Telegram bot token" required:"true"`
	Brokers               string `short:"b" long:"brokers" env:"KAFKA_BROKERS" description:"Kafka brokers list" required:"true"`
	DefaultTopic          string `short:"s" long:"default-topic" env:"DEFAULT_TOPIC" description:"Default topic for messages" required:"false" default:"messages"`
	BrokerConnRetries     int    `short:"r" long:"broker-conn-retries" env:"BROKER_CONC_RETRIES" description:"Number of retries for concurrent broker errors" required:"false" default:"5"`
	BrokerConnRetryTimout int    `short:"d" long:"broker-conn-retry-timeout" env:"BROKER_CONC_RETRY_TIMEOUT" description:"Timeout for broker connection retry in seconds" required:"false" default:"30"`
}

func (o *opts) String() string {
	return fmt.Sprintf(`Brokers: %s
		Default topic: %s
		Broker connection retries: %d
		Broker connection retry timeout: %d`,
		o.Brokers, o.DefaultTopic, o.BrokerConnRetries, o.BrokerConnRetryTimout)
}

func main() {

	var o opts
	if _, err := flags.Parse(&o); err != nil {
		log.Printf("error while parsing flags: %s", err)
		os.Exit(1)
	}

	log.Printf("starting bot with options: %s", o.String())

	logger := setupLogger()

	settings := botSettings(o)

	bot, err := tele.NewBot(settings)

	if err != nil {
		logger.Error(err, "error while creating bot")
		os.Exit(1)
	}

	producer := setupProducer(o, logger)

	sender := &KafkaSender{
		Dest:     setupTextDestination(o),
		Producer: producer,
	}

	txtMsgPipe := Pipe{
		Filters: []Filter{},
		Handler: &TextMessageHandler{},
		Sender:  sender,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		logger.Info("interrupt signal received")
		cancel()
	}()

	g, errCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		bot.Handle(tele.OnText, setupPipeHandler(errCtx, txtMsgPipe, logger))
		bot.Start()
		return nil
	})

	g.Go(func() error {
		<-errCtx.Done()
		logger.Info("context done")
		bot.Stop()
		logger.Info("bot stopped")
		if err := producer.Close(); err != nil {
			return err
		}
		logger.Info("producer closed")
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Info("exit reason: %s \n", err)
	}

}

func setupPipeHandler(ctx context.Context, pipe Pipe, logger logr.Logger) func(ctx tele.Context) error {
	return func(msgCtx tele.Context) error {
		err := pipe.Process(ctx, msgCtx)
		if err == ErrEmptyDestination {
			logger.Error(err, "EmptyDestination error while processing message")
		} else if err != nil {
			logger.Info("error while processing message", "error", err)
		}
		return nil
	}
}

func botSettings(o opts) tele.Settings {
	return tele.Settings{
		Token: o.Token,
		Poller: &tele.LongPoller{
			Timeout:        10 * time.Second,
			AllowedUpdates: []string{"message", "edited_message"},
		},
	}
}

func setupTextDestination(o opts) Destination {
	return &TopicDestination{
		Value: o.DefaultTopic,
	}
}

func parseBrokers(brokers string) []string {
	splitted := strings.Split(brokers, ",")
	for i, broker := range splitted {
		splitted[i] = strings.TrimSpace(broker)
	}
	return splitted
}

func setupProducer(o opts, logger logr.Logger) sarama.AsyncProducer {
	brokerList := parseBrokers(o.Brokers)

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	var producer sarama.AsyncProducer
	var err error
	for i := 0; i < o.BrokerConnRetries; i++ {
		producer, err = sarama.NewAsyncProducer(brokerList, config)

		if err != nil {
			logger.Info("error while creating kafka producer", "error", err, "attempts left", o.BrokerConnRetries-i)
			logger.Info("retrying in", "timeout", o.BrokerConnRetryTimout, "seconds")
			<-time.After(time.Duration(o.BrokerConnRetryTimout) * time.Second)
		}
	}

	if producer == nil {
		logger.Error(err, "error while creating kafka producer")
		os.Exit(1)
	}

	go func() {
		for err := range producer.Errors() {
			logger.Info("error while sending message to kafka", "error", err)
		}
	}()

	return producer
}

func setupLogger() logr.Logger {
	return stdoutLogger()
}

func stdoutLogger() logr.Logger {
	zl := zerolog.New(os.Stdout)
	zl = zl.With().Caller().Timestamp().Logger()
	return zerologr.New(&zl)
}
