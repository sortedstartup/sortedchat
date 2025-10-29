package dao

type DAO interface {
	Infer(dummy string) error
	CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error)
	ListProducts() ([]*Product, error)
	GetProductById(productID string) (*Product, error)

	// Subscription methods
	CreateSubscription(userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) (string, error)
	UpdateSubscription(subscriptionID, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) error
	GetSubscriptionByProviderCustomerID(providerCustomerID string) (*Subscription, error)
	GetSubscriptionByUserIDAndProductID(userID, productID string) (*Subscription, error)

	// User payment methods
	CreateUserPayment(userID, productID, subscriptionID, paymentID, transactionMetadata string, isSuccess bool) (string, error)
}
