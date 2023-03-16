package main

import (
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"

	tele "gopkg.in/telebot.v3"
)

func main() {

	log.Print("started timeout for 1 min")

	<- time.After(1 * time.Minute)

	log.Print("timeout finished, setting up bot...")

	token := os.Getenv("TG_BOT_TOKEN")
	if token == "" {
		log.Fatal("TG_BOT_TOKEN is not set")
		return
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		log.Fatal("KAFKA_BROKER is not set")
		return
	}

	settings := tele.Settings{
		Token:  os.Getenv("TG_BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(settings)

	if err != nil {
		log.Fatal(err)
		return
	}

	txtMsgPipe := setupPipe([]string{broker})

	bot.Handle(tele.OnText, func(ctx tele.Context) error {
		if err = txtMsgPipe.Process(ctx, nil); err != nil {
			log.Print(err)
		}
		return nil
	})

	bot.Start()
}

func setupHandler() Handler {
	return &TextMessageHandler{}
}

func setupTextDestination() Destination {
	return &TopicDestination{
		Value: "textmessage",
	}
}

func setupSender(brokerList []string) Sender {
	return &KafkaSender{
		Dest: setupTextDestination(),
		Producer: setupProducer(brokerList),
	}
}

func setupProducer(brokerList []string) sarama.AsyncProducer {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewAsyncProducer(brokerList, config)

	if err != nil {
		log.Fatalln("Failed to start kafka producer:", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Println("Failed to write access log entry:", err)
		}
	}()

	return producer
}

func setupPipe(brokerList []string) Pipe {
	return Pipe{
		Filters: []Filter{},
		Handler: setupHandler(),
		Sender: setupSender(brokerList),
	}
}