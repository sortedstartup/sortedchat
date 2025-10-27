package dao

type DAO interface {
	Infer(dummy string) error
	CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error)
	ListProducts() ([]*Product, error)
	CreateUserPurchase(sessionID string, userID string, productID string, transaction_metadata string, is_success bool, provider string) (string, error)
	GetProductById(productID string) (*Product, error)

	// Subscription methods
	CreateSubscription(userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) (string, error)
	UpdateSubscription(subscriptionID, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) error
	GetSubscriptionByID(subscriptionID string) (*Subscription, error)
	GetSubscriptionByProviderID(providerCustomerID string) (*Subscription, error)
	CheckUserProductAccess(userID, productID string) (*Subscription, error)

	// User payment methods
	CreateUserPayment(userID, productID, subscriptionID, paymentID, transactionMetadata string) (string, error)
}
