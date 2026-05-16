// Package domain defines core interfaces (Ports) and domain types for
// ViaBot. It has zero external dependencies.
package domain

import "context"

// ---------------------------------------------------------------------------
// Port: Messenger
// Abstraction for bidirectional message exchange across any channel
// (WhatsApp, Telegram, Email, SMS, etc.).
// ---------------------------------------------------------------------------

// Channel represents a messaging platform.
type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelTelegram Channel = "telegram"
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
)

// CustomerID uniquely identifies a customer across any channel.
type CustomerID struct {
	value string
}

// NewCustomerID creates a new CustomerID.
func NewCustomerID(value string) CustomerID {
	return CustomerID{value: value}
}

// Value returns the CustomerID string (JID for WhatsApp, chat ID for Telegram, etc.).
func (id CustomerID) Value() string {
	return id.value
}

// Recipient encapsulates where and how to deliver a message.
type Recipient struct {
	id      CustomerID
	channel Channel
}

// NewRecipient creates a new Recipient.
func NewRecipient(id CustomerID, channel Channel) Recipient {
	return Recipient{id: id, channel: channel}
}

// MediaAttachment holds optional media content for a message.
type MediaAttachment struct {
	mimeType string
	data     []byte
	filename string
}

// OutboundMessage is a message sent TO a customer.
type OutboundMessage struct {
	text  string
	media *MediaAttachment
}

// InboundMessage is a message received FROM a customer.
type InboundMessage struct {
	from     CustomerID
	text     string
	media    *MediaAttachment
	received interface{} // timestamp, typed appropriately at adapter level
}

// MessageHandler processes incoming messages.
type MessageHandler func(ctx context.Context, msg *InboundMessage) error

// MessageSender is the ability to send messages to customers.
// Segregated from MessageReceiver for ISP compliance.
type MessageSender interface {
	Send(ctx context.Context, to Recipient, message *OutboundMessage) error
}

// MessageReceiver is the ability to receive messages from customers.
type MessageReceiver interface {
	Receive(ctx context.Context, handler MessageHandler) error
}

// Messenger combines both sending and receiving.
type Messenger interface {
	MessageSender
	MessageReceiver
}
