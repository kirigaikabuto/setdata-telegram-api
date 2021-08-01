package main

import (
	"fmt"
	"github.com/djumanoff/amqp"
	"github.com/joho/godotenv"
	setdata_common "github.com/kirigaikabuto/setdata-common"
	setdata_telegram_store "github.com/kirigaikabuto/setdata-telegram-store"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

var (
	configPath           = ".env"
	version              = "0.0.0"
	amqpHost             = ""
	amqpPort             = 0
	postgresUser         = ""
	postgresPassword     = ""
	postgresDatabaseName = ""
	postgresHost         = ""
	postgresPort         = 5432
	postgresParams       = ""
	flags                = []cli.Flag{
		&cli.StringFlag{
			Name:        "config, c",
			Usage:       "path to .env config file",
			Destination: &configPath,
		},
	}
)

func parseEnvFile() {
	// Parse config file (.env) if path to it specified and populate env vars
	if configPath != "" {
		godotenv.Overload(configPath)
	}
	amqpHost = os.Getenv("RABBIT_HOST")
	amqpPortStr := os.Getenv("RABBIT_PORT")
	amqpPort, _ = strconv.Atoi(amqpPortStr)
	if amqpPort == 0 {
		amqpPort = 5672
	}
	if amqpHost == "" {
		amqpHost = "localhost"
	}
	postgresUser = os.Getenv("POSTGRES_USER")
	postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	postgresDatabaseName = os.Getenv("POSTGRES_DATABASE")
	postgresParams = os.Getenv("POSTGRES_PARAMS")
	portStr := os.Getenv("POSTGRES_PORT")
	postgresPort, _ = strconv.Atoi(portStr)
	postgresHost = os.Getenv("POSTGRES_HOST")
	if postgresHost == "" {
		postgresHost = "localhost"
	}
	if postgresPort == 0 {
		postgresPort = 5432
	}
	if postgresUser == "" {
		postgresUser = "setdatauser"
	}
	if postgresPassword == "" {
		postgresPassword = "123456789"
	}
	if postgresDatabaseName == "" {
		postgresDatabaseName = "setdata"
	}
	if postgresParams == "" {
		postgresParams = "sslmode=disable"
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Set Data Acl Store Api"
	app.Description = ""
	app.Usage = "set data run"
	app.UsageText = "set data run"
	app.Version = version
	app.Flags = flags
	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

func run(c *cli.Context) error {
	parseEnvFile()
	fmt.Println(postgresHost)
	fmt.Println(postgresUser)
	fmt.Println(postgresDatabaseName)
	fmt.Println(postgresPassword)
	fmt.Println(postgresParams)
	rabbitConfig := amqp.Config{
		Host:     amqpHost,
		Port:     amqpPort,
		LogLevel: 5,
	}
	serverConfig := amqp.ServerConfig{
		ResponseX: "response",
		RequestX:  "request",
	}
	sess := amqp.NewSession(rabbitConfig)
	err := sess.Connect()
	if err != nil {
		return err
	}
	srv, err := sess.Server(serverConfig)
	if err != nil {
		return err
	}
	config := setdata_telegram_store.PostgresConfig{
		Host:             postgresHost,
		Port:             postgresPort,
		User:             postgresUser,
		Password:         postgresPassword,
		Database:         postgresDatabaseName,
		Params:           postgresParams,
		ConnectionString: "",
	}
	telegramPostgresStore, err := setdata_telegram_store.NewPostgresTelegramStore(config)
	if err != nil {
		return err
	}
	chatIdPostgresStore, err := setdata_telegram_store.NewPostgresChatIdStore(config)
	if err != nil {
		return err
	}
	telegramStoreService := setdata_telegram_store.NewTelegramService("123456789", telegramPostgresStore, chatIdPostgresStore)
	telegramAmqpEndpoints := setdata_telegram_store.NewTelegramAmqpEndpoints(setdata_common.NewCommandHandler(telegramStoreService))
	srv.Endpoint("telegram.sendMessage", telegramAmqpEndpoints.SendMessageTelegramBotAmqpEndpoint())
	srv.Endpoint("telegram.create", telegramAmqpEndpoints.CreateTelegramBotAmqpEndpoint())
	srv.Endpoint("telegram.list", telegramAmqpEndpoints.ListTelegramBotAmqpEndpoint())
	err = srv.Start()
	if err != nil {
		return err
	}
	return nil
}
