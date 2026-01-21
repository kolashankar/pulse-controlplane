package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"pulse-control-plane/models"

	razorpay "github.com/razorpay/razorpay-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// RazorpayService handles Razorpay payment operations
type RazorpayService struct {
	db             *mongo.Database
	client         *razorpay.Client
	keyID          string
	keySecret      string
	webhookSecret  string
	billingService *BillingService
}

// NewRazorpayService creates a new Razorpay service
func NewRazorpayService(db *mongo.Database, keyID, keySecret, webhookSecret string, billingService *BillingService) *RazorpayService {
	client := razorpay.NewClient(keyID, keySecret)
	
	return &RazorpayService{
		db:             db,
		client:         client,
		keyID:          keyID,
		keySecret:      keySecret,
		webhookSecret:  webhookSecret,
		billingService: billingService,
	}
}

// RazorpayCustomer represents a Razorpay customer
type RazorpayCustomer struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id"`
	CustomerID     string             `bson:"customer_id" json:"customer_id"`
	Email          string             `bson:"email" json:"email"`
	Name           string             `bson:"name" json:"name"`
	Contact        string             `bson:"contact" json:"contact"`
	GSTNumber      string             `bson:"gst_number,omitempty" json:"gst_number,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

func (RazorpayCustomer) TableName() string {
	return "razorpay_customers"
}

// RazorpaySubscription represents a Razorpay subscription
type RazorpaySubscription struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID        primitive.ObjectID `bson:"project_id" json:"project_id"`
	OrgID            primitive.ObjectID `bson:"org_id" json:"org_id"`
	SubscriptionID   string             `bson:"subscription_id" json:"subscription_id"`
	PlanID           string             `bson:"plan_id" json:"plan_id"`
	CustomerID       string             `bson:"customer_id" json:"customer_id"`
	Status           string             `bson:"status" json:"status"` // created, authenticated, active, paused, cancelled
	CurrentStart     time.Time          `bson:"current_start" json:"current_start"`
	CurrentEnd       time.Time          `bson:"current_end" json:"current_end"`
	ChargeAt         time.Time          `bson:"charge_at" json:"charge_at"`
	PaidCount        int                `bson:"paid_count" json:"paid_count"`
	RemainingCount   int                `bson:"remaining_count" json:"remaining_count"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

func (RazorpaySubscription) TableName() string {
	return "razorpay_subscriptions"
}

// RazorpayPayment represents a Razorpay payment
type RazorpayPayment struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID      primitive.ObjectID `bson:"project_id" json:"project_id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id"`
	PaymentID      string             `bson:"payment_id" json:"payment_id"`
	OrderID        string             `bson:"order_id" json:"order_id"`
	InvoiceID      primitive.ObjectID `bson:"invoice_id,omitempty" json:"invoice_id,omitempty"`
	Amount         float64            `bson:"amount" json:"amount"`
	Currency       string             `bson:"currency" json:"currency"`
	Status         string             `bson:"status" json:"status"` // created, authorized, captured, refunded, failed
	Method         string             `bson:"method" json:"method"` // card, netbanking, wallet, upi
	Email          string             `bson:"email" json:"email"`
	Contact        string             `bson:"contact" json:"contact"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

func (RazorpayPayment) TableName() string {
	return "razorpay_payments"
}

// CreateCustomer creates a new customer in Razorpay
func (s *RazorpayService) CreateCustomer(ctx context.Context, orgID primitive.ObjectID, name, email, contact, gstNumber string) (*RazorpayCustomer, error) {
	// Create customer in Razorpay
	data := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	
	if contact != "" {
		data["contact"] = contact
	}
	
	if gstNumber != "" {
		data["gstin"] = gstNumber
	}

	body, err := s.client.Customer.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay customer: %w", err)
	}

	customerID, ok := body["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid customer ID response from Razorpay")
	}

	// Store customer in database
	customer := &RazorpayCustomer{
		ID:         primitive.NewObjectID(),
		OrgID:      orgID,
		CustomerID: customerID,
		Email:      email,
		Name:       name,
		Contact:    contact,
		GSTNumber:  gstNumber,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	collection := s.db.Collection(customer.TableName())
	_, err = collection.InsertOne(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("failed to store customer: %w", err)
	}

	return customer, nil
}

// CreateSubscription creates a new subscription in Razorpay
func (s *RazorpayService) CreateSubscription(ctx context.Context, projectID, orgID primitive.ObjectID, planID, customerID string, totalCount int) (*RazorpaySubscription, error) {
	// Create subscription in Razorpay
	data := map[string]interface{}{
		"plan_id":      planID,
		"customer_id":  customerID,
		"total_count":  totalCount,
		"quantity":     1,
		"start_at":     time.Now().Add(24 * time.Hour).Unix(), // Start tomorrow
		"notify":       1,
	}

	body, err := s.client.Subscription.Create(data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay subscription: %w", err)
	}

	subscriptionID, ok := body["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid subscription ID response from Razorpay")
	}

	status, _ := body["status"].(string)

	// Store subscription in database
	subscription := &RazorpaySubscription{
		ID:             primitive.NewObjectID(),
		ProjectID:      projectID,
		OrgID:          orgID,
		SubscriptionID: subscriptionID,
		PlanID:         planID,
		CustomerID:     customerID,
		Status:         status,
		CurrentStart:   time.Now(),
		CurrentEnd:     time.Now().AddDate(0, 1, 0), // Monthly
		ChargeAt:       time.Now().Add(24 * time.Hour),
		PaidCount:      0,
		RemainingCount: totalCount,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	collection := s.db.Collection(subscription.TableName())
	_, err = collection.InsertOne(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to store subscription: %w", err)
	}

	return subscription, nil
}

// GeneratePaymentLink creates a payment link for an invoice
func (s *RazorpayService) GeneratePaymentLink(ctx context.Context, invoiceID primitive.ObjectID) (string, error) {
	// Get invoice
	invoice, err := s.billingService.GetInvoice(ctx, invoiceID)
	if err != nil {
		return "", fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get organization
	orgCollection := s.db.Collection(models.Organization{}.TableName())
	var org models.Organization
	err = orgCollection.FindOne(ctx, bson.M{"_id": invoice.OrgID}).Decode(&org)
	if err != nil {
		return "", fmt.Errorf("failed to find organization: %w", err)
	}

	// Create payment link
	data := map[string]interface{}{
		"amount":      int(invoice.Total * 100), // Convert to paise
		"currency":    "INR",
		"description": fmt.Sprintf("Invoice %s", invoice.InvoiceNumber),
		"customer": map[string]interface{}{
			"name":  org.Name,
			"email": org.AdminEmail,
		},
		"notify": map[string]interface{}{
			"sms":   true,
			"email": true,
		},
		"reminder_enable": true,
		"callback_url":    fmt.Sprintf("https://pulse.io/billing/payment/callback?invoice_id=%s", invoiceID.Hex()),
		"callback_method": "get",
	}

	body, err := s.client.PaymentLink.Create(data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create payment link: %w", err)
	}

	shortURL, ok := body["short_url"].(string)
	if !ok {
		return "", fmt.Errorf("invalid payment link response from Razorpay")
	}

	// Update invoice with payment link
	invoiceCollection := s.db.Collection(models.Invoice{}.TableName())
	_, err = invoiceCollection.UpdateOne(
		ctx,
		bson.M{"_id": invoiceID},
		bson.M{"$set": bson.M{
			"payment_link": shortURL,
			"updated_at":   time.Now(),
		}},
	)
	if err != nil {
		return "", fmt.Errorf("failed to update invoice: %w", err)
	}

	return shortURL, nil
}

// VerifyPayment verifies a payment using signature
func (s *RazorpayService) VerifyPayment(ctx context.Context, paymentID, orderID, signature string) (bool, error) {
	// Verify signature
	message := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(s.keySecret))
	mac.Write([]byte(message))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if expectedSignature != signature {
		return false, fmt.Errorf("invalid payment signature")
	}

	// Fetch payment details from Razorpay
	body, err := s.client.Payment.Fetch(paymentID, nil, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch payment: %w", err)
	}

	status, _ := body["status"].(string)
	if status != "captured" && status != "authorized" {
		return false, fmt.Errorf("payment not successful: %s", status)
	}

	return true, nil
}

// ProcessWebhook processes Razorpay webhook events
func (s *RazorpayService) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	if !s.verifyWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := bson.UnmarshalExtJSON(payload, false, &webhookData); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	event, ok := webhookData["event"].(string)
	if !ok {
		return fmt.Errorf("invalid webhook event")
	}

	// Handle different webhook events
	switch event {
	case "payment.captured":
		return s.handlePaymentCaptured(ctx, webhookData)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, webhookData)
	case "subscription.charged":
		return s.handleSubscriptionCharged(ctx, webhookData)
	case "subscription.cancelled":
		return s.handleSubscriptionCancelled(ctx, webhookData)
	case "subscription.paused":
		return s.handleSubscriptionPaused(ctx, webhookData)
	case "subscription.resumed":
		return s.handleSubscriptionResumed(ctx, webhookData)
	default:
		// Unknown event, log and ignore
		return nil
	}
}

// verifyWebhookSignature verifies the webhook signature
func (s *RazorpayService) verifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return expectedSignature == signature
}

// handlePaymentCaptured handles payment captured events
func (s *RazorpayService) handlePaymentCaptured(ctx context.Context, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload")
	}

	payment, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment data")
	}

	paymentID, _ := payment["id"].(string)
	orderID, _ := payment["order_id"].(string)
	amount, _ := payment["amount"].(float64)

	// Store payment record
	paymentRecord := &RazorpayPayment{
		ID:        primitive.NewObjectID(),
		PaymentID: paymentID,
		OrderID:   orderID,
		Amount:    amount / 100, // Convert from paise
		Currency:  "INR",
		Status:    "captured",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := s.db.Collection(paymentRecord.TableName())
	_, err := collection.InsertOne(ctx, paymentRecord)
	if err != nil {
		return fmt.Errorf("failed to store payment: %w", err)
	}

	return nil
}

// handlePaymentFailed handles payment failed events
func (s *RazorpayService) handlePaymentFailed(ctx context.Context, data map[string]interface{}) error {
	// Similar implementation to handlePaymentCaptured but mark as failed
	return nil
}

// handleSubscriptionCharged handles subscription charged events
func (s *RazorpayService) handleSubscriptionCharged(ctx context.Context, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload")
	}

	subscription, ok := payload["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}

	subscriptionID, _ := subscription["id"].(string)

	// Update subscription in database
	collection := s.db.Collection(RazorpaySubscription{}.TableName())
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"subscription_id": subscriptionID},
		bson.M{
			"$set": bson.M{
				"status":     "active",
				"updated_at": time.Now(),
			},
			"$inc": bson.M{
				"paid_count": 1,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// handleSubscriptionCancelled handles subscription cancelled events
func (s *RazorpayService) handleSubscriptionCancelled(ctx context.Context, data map[string]interface{}) error {
	payload, ok := data["payload"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload")
	}

	subscription, ok := payload["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data")
	}

	subscriptionID, _ := subscription["id"].(string)

	// Update subscription status
	collection := s.db.Collection(RazorpaySubscription{}.TableName())
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"subscription_id": subscriptionID},
		bson.M{"$set": bson.M{
			"status":     "cancelled",
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	return nil
}

// handleSubscriptionPaused handles subscription paused events
func (s *RazorpayService) handleSubscriptionPaused(ctx context.Context, data map[string]interface{}) error {
	// Similar to handleSubscriptionCancelled
	return nil
}

// handleSubscriptionResumed handles subscription resumed events
func (s *RazorpayService) handleSubscriptionResumed(ctx context.Context, data map[string]interface{}) error {
	// Similar to handleSubscriptionCharged
	return nil
}

// GetInvoices retrieves Razorpay invoices for a project
func (s *RazorpayService) GetInvoices(ctx context.Context, projectID primitive.ObjectID, page, limit int) ([]models.Invoice, int64, error) {
	return s.billingService.ListInvoices(ctx, projectID, page, limit)
}

// ProcessRefund processes a refund for a payment
func (s *RazorpayService) ProcessRefund(ctx context.Context, paymentID string, amount float64) (string, error) {
	// Create refund in Razorpay
	data := map[string]interface{}{
		"amount": int(amount * 100), // Convert to paise
	}

	body, err := s.client.Payment.Refund(paymentID, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refund: %w", err)
	}

	refundID, ok := body["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid refund ID response from Razorpay")
	}

	// Update payment status
	collection := s.db.Collection(RazorpayPayment{}.TableName())
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"payment_id": paymentID},
		bson.M{"$set": bson.M{
			"status":     "refunded",
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		return "", fmt.Errorf("failed to update payment status: %w", err)
	}

	return refundID, nil
}

// GetCustomerByOrgID retrieves a customer by organization ID
func (s *RazorpayService) GetCustomerByOrgID(ctx context.Context, orgID primitive.ObjectID) (*RazorpayCustomer, error) {
	collection := s.db.Collection(RazorpayCustomer{}.TableName())
	var customer RazorpayCustomer
	err := collection.FindOne(ctx, bson.M{"org_id": orgID}).Decode(&customer)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetSubscriptionByProjectID retrieves a subscription by project ID
func (s *RazorpayService) GetSubscriptionByProjectID(ctx context.Context, projectID primitive.ObjectID) (*RazorpaySubscription, error) {
	collection := s.db.Collection(RazorpaySubscription{}.TableName())
	var subscription RazorpaySubscription
	err := collection.FindOne(ctx, bson.M{"project_id": projectID}).Decode(&subscription)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}
