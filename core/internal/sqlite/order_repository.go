package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"viabot.stream/sdk/domain"
)

// ---------------------------------------------------------------------------
// SQLite adapter for OrderRepository
// Rule 8: 1 instance variable (db).
// ---------------------------------------------------------------------------

// OrderRepository implements domain.OrderRepository using SQLite.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new SQLite-backed OrderRepository.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Save persists an order and its items within a transaction.
func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() // no-op after commit

	if err := r.upsertOrder(ctx, tx, order); err != nil {
		return fmt.Errorf("upsert order: %w", err)
	}
	if err := r.replaceOrderItems(ctx, tx, order); err != nil {
		return fmt.Errorf("replace items: %w", err)
	}
	return tx.Commit()
}

// FindByID retrieves a full Order aggregate by ID.
func (r *OrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	order, err := r.scanOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	items, err := r.scanOrderItems(ctx, id, order.currency)
	if err != nil {
		return nil, err
	}
	orderDetails := domain.NewOrderItemCollection(items)
	return r.reconstructOrder(order, orderDetails)
}

// FindByCustomer returns all orders for a given customer.
func (r *OrderRepository) FindByCustomer(ctx context.Context, customerID domain.CustomerID) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, customer_id, status, total_amount, currency,
		        COALESCE(payment_id,''), COALESCE(payment_status,''), COALESCE(payment_method,''),
		        COALESCE(paid_amount,0), COALESCE(checkout_link,''), COALESCE(qr_code,''),
		        COALESCE(qr_code_text,''), COALESCE(boleto_barcode,''), COALESCE(boleto_url,''),
		        COALESCE(payment_expires_at,''), COALESCE(payment_created_at,''),
		        COALESCE(payment_confirmed_at,''),
		        created_at, COALESCE(confirmed_at,'')
		 FROM orders WHERE customer_id = ? ORDER BY created_at DESC`, customerID.Value())
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order, err := r.scanOrderFromRow(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return orders, nil
}

// FindPendingByCustomer finds a single pending order for the customer.
func (r *OrderRepository) FindPendingByCustomer(ctx context.Context, customerID domain.CustomerID) (*domain.Order, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM orders WHERE customer_id = ? AND status = 'pending' LIMIT 1`,
		customerID.Value()).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find pending: %w", err)
	}
	return r.FindByID(ctx, domain.NewOrderID(id))
}

// UpdateStatus changes the order status.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id domain.OrderID, status domain.OrderStatus) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = ? WHERE id = ?`, string(status), id.Value())
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// paymentFields groups serialized payment columns for the orders table.
type paymentFields struct {
	id, status, method, checkoutLink, qrCode, qrCodeText  string
	boletoBarcode, boletoURL                               string
	paidAmount                                             int64
	expiresAt, createdAt, confirmedAt                      string
}

// serializePayment extracts payment data for SQL persistence.
func serializePayment(order *domain.Order) paymentFields {
	var f paymentFields
	if p := order.Payment(); p != nil {
		f.id = p.ID()
		f.status = string(p.Status())
		f.method = string(p.Method())
		f.checkoutLink = p.CheckoutLink()
		f.qrCode = p.QRCode()
		f.qrCodeText = p.QRCodeText()
		f.boletoBarcode = p.BoletoBarcode()
		f.boletoURL = p.BoletoURL()
		f.paidAmount = p.PaidAmount().Amount()
		if !p.ExpiresAt().IsZero() {
			f.expiresAt = p.ExpiresAt().Format(time.RFC3339)
		}
		if !p.CreatedAt().IsZero() {
			f.createdAt = p.CreatedAt().Format(time.RFC3339)
		}
		if !p.ConfirmedAt().IsZero() {
			f.confirmedAt = p.ConfirmedAt().Format(time.RFC3339)
		}
	}
	return f
}

// orderConfirmedAt returns the formatted confirmed_at or empty.
func orderConfirmedAt(order *domain.Order) string {
	if order.ConfirmedAt().IsZero() {
		return ""
	}
	return order.ConfirmedAt().Format(time.RFC3339)
}

func (r *OrderRepository) upsertOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	f := serializePayment(order)

	_, err := tx.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, total_amount, currency,
		                    payment_id, payment_status, payment_method, paid_amount,
		                    checkout_link, qr_code, qr_code_text,
		                    boleto_barcode, boleto_url,
		                    payment_expires_at, payment_created_at, payment_confirmed_at,
		                    created_at, confirmed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status, payment_id=excluded.payment_id,
			payment_status=excluded.payment_status, payment_method=excluded.payment_method,
			paid_amount=excluded.paid_amount, checkout_link=excluded.checkout_link,
			qr_code=excluded.qr_code, qr_code_text=excluded.qr_code_text,
			boleto_barcode=excluded.boleto_barcode, boleto_url=excluded.boleto_url,
			payment_expires_at=excluded.payment_expires_at,
			payment_created_at=excluded.payment_created_at,
			payment_confirmed_at=excluded.payment_confirmed_at,
			confirmed_at=excluded.confirmed_at`,
		order.ID().Value(), order.CustomerID().Value(), string(order.Status()),
		order.Total().Amount(), order.Total().Currency(),
		nullIfEmpty(f.id), nullIfEmpty(f.status), nullIfEmpty(f.method), f.paidAmount,
		nullIfEmpty(f.checkoutLink), nullIfEmpty(f.qrCode), nullIfEmpty(f.qrCodeText),
		nullIfEmpty(f.boletoBarcode), nullIfEmpty(f.boletoURL),
		nullIfEmpty(f.expiresAt), nullIfEmpty(f.createdAt), nullIfEmpty(f.confirmedAt),
		order.CreatedAt().Format(time.RFC3339), nullIfEmpty(orderConfirmedAt(order)))
	return err
}

func (r *OrderRepository) replaceOrderItems(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	// Delete existing items then re-insert
	if _, err := tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = ?`, order.ID().Value()); err != nil {
		return err
	}
	for _, item := range order.Items().Items() {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, name, quantity, unit_price)
			VALUES (?, ?, ?, ?, ?)`,
			order.ID().Value(), item.ProductID().Value(), item.Name(),
			item.Quantity(), item.UnitPrice().Amount()); err != nil {
			return err
		}
	}
	return nil
}

// orderRow holds scanned columns from the orders table.
type orderRow struct {
	id, customerID, status               string
	totalAmount                          int64
	currency                             string
	paymentID, paymentStatus, paymentMethod string
	paidAmount                           int64
	checkoutLink, qrCode, qrCodeText     string
	boletoBarcode, boletoURL             string
	paymentExpiresAt, paymentCreatedAt, paymentConfirmedAt string
	createdAt, confirmedAt               string
}

func (r *OrderRepository) scanOrder(ctx context.Context, id domain.OrderID) (*orderRow, error) {
	row := &orderRow{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, customer_id, status, total_amount, currency,
		       COALESCE(payment_id,''), COALESCE(payment_status,''), COALESCE(payment_method,''),
		       COALESCE(paid_amount,0), COALESCE(checkout_link,''), COALESCE(qr_code,''),
		       COALESCE(qr_code_text,''), COALESCE(boleto_barcode,''), COALESCE(boleto_url,''),
		       COALESCE(payment_expires_at,''), COALESCE(payment_created_at,''),
		       COALESCE(payment_confirmed_at,''),
		       created_at, COALESCE(confirmed_at,'')
		 FROM orders WHERE id = ?`, id.Value()).Scan(
		&row.id, &row.customerID, &row.status, &row.totalAmount, &row.currency,
		&row.paymentID, &row.paymentStatus, &row.paymentMethod, &row.paidAmount,
		&row.checkoutLink, &row.qrCode, &row.qrCodeText,
		&row.boletoBarcode, &row.boletoURL,
		&row.paymentExpiresAt, &row.paymentCreatedAt, &row.paymentConfirmedAt,
		&row.createdAt, &row.confirmedAt)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan order: %w", err)
	}
	return row, nil
}

func (r *OrderRepository) scanOrderItems(ctx context.Context, id domain.OrderID, currency string) ([]domain.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT product_id, name, unit_price, quantity FROM order_items WHERE order_id = ?`, id.Value())
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()
	return r.scanOrderItemRows(rows, currency)
}

// scanOrderItemRows reads item rows from the query result.
func (r *OrderRepository) scanOrderItemRows(rows *sql.Rows, currency string) ([]domain.OrderItem, error) {
	var items []domain.OrderItem
	for rows.Next() {
		var productID, name string
		var unitPrice, quantity int
		if err := rows.Scan(&productID, &name, &unitPrice, &quantity); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		price, err := domain.NewMoney(int64(unitPrice), currency)
		if err != nil {
			return nil, fmt.Errorf("invalid item price: %w", err)
		}
		item, err := domain.NewOrderItem(domain.NewProductID(productID), name, price, quantity)
		if err != nil {
			return nil, fmt.Errorf("reconstruct item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *OrderRepository) scanOrderFromRow(row interface{ Scan(...interface{}) error }) (*domain.Order, error) {
	o := &orderRow{}
	err := row.Scan(
		&o.id, &o.customerID, &o.status, &o.totalAmount, &o.currency,
		&o.paymentID, &o.paymentStatus, &o.paymentMethod, &o.paidAmount,
		&o.checkoutLink, &o.qrCode, &o.qrCodeText,
		&o.boletoBarcode, &o.boletoURL,
		&o.paymentExpiresAt, &o.paymentCreatedAt, &o.paymentConfirmedAt,
		&o.createdAt, &o.confirmedAt)
	if err != nil {
		return nil, fmt.Errorf("scan order row: %w", err)
	}
	return r.reconstructOrderFromRow(o)
}

func (r *OrderRepository) reconstructOrder(row *orderRow, items domain.OrderItemCollection) (*domain.Order, error) {
	customerID := domain.NewCustomerID(row.customerID)
	total, _ := domain.NewMoney(row.totalAmount, row.currency)
	createdAt, _ := time.Parse(time.RFC3339, row.createdAt)

	var confirmedAt *time.Time
	if row.confirmedAt != "" {
		if t, err := time.Parse(time.RFC3339, row.confirmedAt); err == nil {
			confirmedAt = &t
		}
	}

	payment := r.reconstructPayment(row)
	return domain.ReconstructOrder(
		domain.NewOrderID(row.id),
		customerID,
		domain.OrderStatus(row.status),
		items,
		total,
		payment,
		createdAt,
		confirmedAt,
	), nil
}

func (r *OrderRepository) reconstructOrderFromRow(o *orderRow) (*domain.Order, error) {
	// Load items from the database for this order
	items, err := r.scanOrderItems(context.Background(), domain.NewOrderID(o.id), o.currency)
	if err != nil {
		return nil, err
	}
	collection := domain.NewOrderItemCollection(items)
	return r.reconstructOrder(o, collection)
}

func (r *OrderRepository) reconstructPayment(o *orderRow) *domain.Payment {
	if o.paymentID == "" {
		return nil
	}
	amount, _ := domain.NewMoney(o.totalAmount, o.currency)
	paidAmount, _ := domain.NewMoney(o.paidAmount, o.currency)
	expiresAt, _ := time.Parse(time.RFC3339, o.paymentExpiresAt)
	createdAt, _ := time.Parse(time.RFC3339, o.paymentCreatedAt)
	confirmedAt, _ := time.Parse(time.RFC3339, o.paymentConfirmedAt)

	return domain.ReconstructPayment(
		o.paymentID,
		domain.NewOrderID(o.id),
		domain.PaymentStatus(o.paymentStatus),
		amount,
		paidAmount,
		domain.PaymentMethod(o.paymentMethod),
		o.checkoutLink, o.qrCode, o.qrCodeText,
		o.boletoBarcode, o.boletoURL,
		expiresAt, createdAt, confirmedAt,
	)
}

// nullIfEmpty returns nil for empty strings (maps to SQL NULL).
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
