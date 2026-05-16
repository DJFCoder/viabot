package domain

import "time"

// ---------------------------------------------------------------------------
// Customer — Entity
// Rule 8 compliance: Customer composes exactly 2 instance variables.
// ---------------------------------------------------------------------------

// customerProfile groups profile information about a customer.
type customerProfile struct {
	name      string
	phone     PhoneNumber
	createdAt time.Time
}

// Customer represents a person who interacts with the bot.
// Identity is derived from the WhatsApp JID.
type Customer struct {
	id      CustomerID
	profile customerProfile
}

// NewCustomer creates a Customer from first interaction data.
func NewCustomer(id CustomerID, phone PhoneNumber, name string) *Customer {
	return &Customer{
		id: id,
		profile: customerProfile{
			name:      name,
			phone:     phone,
			createdAt: time.Now(),
		},
	}
}

// UpdateName changes the customer's display name.
func (c *Customer) UpdateName(name string) {
	c.profile.name = name
}

// ---------------------------------------------------------------------------
// Read accessors
// ---------------------------------------------------------------------------

// ID returns the unique customer identifier.
func (c *Customer) ID() CustomerID { return c.id }

// Phone returns the customer's phone number.
func (c *Customer) Phone() PhoneNumber { return c.profile.phone }

// Name returns the customer's display name.
func (c *Customer) Name() string { return c.profile.name }

// CreatedAt returns when the customer was first registered.
func (c *Customer) CreatedAt() time.Time { return c.profile.createdAt }
