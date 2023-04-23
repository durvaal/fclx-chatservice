package entity

import (
	"errors"

	"github.com/google/uuid"
)

type ChatConfig struct {
	Model *Model
	// the fields below are mapped the according to the documentation: https://platform.openai.com/docs/api-reference/completions/create
	Temperature      float32
	TopP             float32
	N                int
	Stop             []string
	MaxTokens        int
	PresencePenalty  float32
	FrequencyPenalty float32
}

type Chat struct {
	ID                   string
	UserID               string
	InitialSystemMessage *Message
	Messages             []*Message
	ErasedMessage        []*Message
	Status               string
	TokenUsage           int
	Config               *ChatConfig
}

func NewChat(userID string, initialSystemMessage *Message, chatConfig *ChatConfig) (*Chat, error) {
	chat := &Chat{
		ID:                   uuid.New().String(),
		UserID:               userID,
		InitialSystemMessage: initialSystemMessage,
		Status:               "active",
		Config:               chatConfig,
		TokenUsage:           0,
	}

	if err := chat.Validate(); err != nil {
		return nil, err
	}

	chat.AddMessage(initialSystemMessage)

	return chat, nil
}

func (c *Chat) Validate() error {
	if c.UserID == "" {
		return errors.New("invalid user_id, please provide a valid user_id")
	}

	if c.Status != "active" && c.Status != "ended" {
		return errors.New("invalid status, please provide a valid status (active, ended)")
	}

	if c.Config.Temperature < 0 || c.Config.Temperature > 2 {
		return errors.New("invalid temperature, please provide a valid temperature (between 0 and 2)")
	}

	// TODO: add more validation according openai documentation

	return nil
}

func (c *Chat) AddMessage(m *Message) error {
	if c.Status == "endend" {
		return errors.New("chat is ended, no more messages allowed")
	}

	for {
		hasAvailableTokensToAddMessage := c.Config.Model.GetMaxTokens() >= m.GetTokens()+c.TokenUsage

		if hasAvailableTokensToAddMessage {
			c.Messages = append(c.Messages, m)
			c.RefreshTokenUsage()
			break
		}

		c.ErasedMessage = append(c.ErasedMessage, c.Messages[0])
		c.Messages = c.Messages[1:] // 1: slice array from first position
		c.RefreshTokenUsage()
	}

	return nil
}

func (c *Chat) GetMessage() []*Message {
	return c.Messages
}

func (c *Chat) CountMessages() int {
	return len(c.Messages)
}

func (c *Chat) End() {
	c.Status = "ended"
}

func (c *Chat) RefreshTokenUsage() {
	c.TokenUsage = 0

	for m := range c.Messages {
		c.TokenUsage += c.Messages[m].GetTokens()
	}
}
