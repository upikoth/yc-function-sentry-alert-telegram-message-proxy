package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type RequestToTelegram struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

func Handler(ctx context.Context, req json.RawMessage) (*Response, error) {
	// Чтение .env файла нужно только при локальной разработке.
	// В других случаях значения переменных окружения уже должны быть установлены.
	// Поэтому ошибку загрузки файла обрабатывать не нужно.
	_ = godotenv.Load()
	config, err := newConfig()

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Printf("%+v\n", ctx)
	fmt.Printf("%s\n", string(req))

	requestData, err := parseRequest(req)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	title := requestData.Data.Event.Title

	if title == "" {
		title = requestData.Data.Issue.Title
	}

	jsonData, err := json.Marshal(RequestToTelegram{
		ChatID: config.TelegramChatID,
		Text: fmt.Sprintf(
			"%s\n\n%s\n%s\n\n%s",
			requestData.Action,
			title,
			requestData.Data.Event.Message,
			requestData.Data.Event.Url,
		),
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	res, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.TelegramToken),
		"application/json",
		bytes.NewBuffer(jsonData))

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return &Response{
		StatusCode: http.StatusOK,
		Body:       string(body),
	}, nil
}

type Config struct {
	TelegramToken  string `envconfig:"NOTIFICATIONS_TELEGRAM_TOKEN" required:"true"`
	TelegramChatID string `envconfig:"NOTIFICATIONS_TELEGRAM_TO" required:"true"`
}

func newConfig() (*Config, error) {
	config := &Config{}

	if err := envconfig.Process("", config); err != nil {
		return nil, err
	}

	return config, nil
}

type Request struct {
	Body       string `json:"body"`
	HTTPMethod string `json:"httpMethod"`
}

type RequestData struct {
	Action string `json:"action"`
	Data   struct {
		Issue struct {
			Title string `json:"title"`
		} `json:"issue"`
		Event struct {
			Title   string `json:"title"`
			Message string `json:"message"`
			Url     string `json:"web_url"`
		} `json:"event"`
	} `json:"data"`
}

func parseRequest(req json.RawMessage) (*RequestData, error) {
	request := &Request{}
	err := json.Unmarshal(req, request)

	if err != nil {
		return nil, err
	}

	requestBody := &RequestData{}
	err = json.Unmarshal([]byte(request.Body), requestBody)

	if err != nil {
		return nil, err
	}

	return requestBody, nil
}
