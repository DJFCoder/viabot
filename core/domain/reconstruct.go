package domain

import "time"

// ---------------------------------------------------------------------------
// Reconstruction functions — used by repository adapters to rehydrate
// aggregates from persisted data. Unlike constructors, these trust the input
// (data was already validated when saved).
// ---------------------------------------------------------------------------

// ReconstructOrder creates an Order from persisted data.
func ReconstructOrder(
	id OrderID,
	customerID CustomerID,
	status OrderStatus,
	items OrderItemCollection,
	total Money,
	payment *Payment,
	createdAt time.Time,
	confirmedAt *time.Time,
) *Order {
	return &Order{
		id: id,
		details: orderDetails{
			customerID:  customerID,
			status:      status,
			items:       items,
			total:       total,
			payment:     payment,
			createdAt:   createdAt,
			confirmedAt: confirmedAt,
		},
	}
}

// ReconstructPayment creates a Payment from persisted data.
func ReconstructPayment(
	id string,
	orderID OrderID,
	status PaymentStatus,
	amount Money,
	paidAmount Money,
	method PaymentMethod,
	checkoutLink, qrCode, qrCodeText string,
	boletoBarcode, boletoURL string,
	expiresAt, createdAt, confirmedAt time.Time,
) *Payment {
	return &Payment{
		id:            id,
		orderID:       orderID,
		status:        status,
		amount:        amount,
		paidAmount:    paidAmount,
		method:        method,
		checkoutLink:  checkoutLink,
		qrCode:        qrCode,
		qrCodeText:    qrCodeText,
		boletoBarcode: boletoBarcode,
		boletoURL:     boletoURL,
		expiresAt:     expiresAt,
		createdAt:     createdAt,
		confirmedAt:   confirmedAt,
	}
}

// ReconstructCart creates a Cart from persisted data.
func ReconstructCart(id CartID, customerID CustomerID, items CartItemCollection, expiresAt time.Time) *Cart {
	return &Cart{
		id: id,
		contents: cartContents{
			customerID: customerID,
			items:      items,
			expiresAt:  expiresAt,
		},
	}
}

// ReconstructCustomer creates a Customer from persisted data.
func ReconstructCustomer(id CustomerID, phone PhoneNumber, name string, createdAt time.Time) *Customer {
	return &Customer{
		id: id,
		profile: customerProfile{
			name:      name,
			phone:     phone,
			createdAt: createdAt,
		},
	}
}

// ReconstructProduct creates a Product from persisted data.
func ReconstructProduct(
	id ProductID,
	name, description string,
	price Money,
	category string,
	available bool,
	imageURL string,
) *Product {
	return &Product{
		id: id,
		details: productDetails{
			name:        name,
			description: description,
			price:       price,
			category:    category,
			available:   available,
			imageURL:    imageURL,
		},
	}
}
