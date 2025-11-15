package dao

type DAO interface {
	CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error)
	ListProducts(userID string) ([]*Product, error)
	GetProductById(productID string) (*Product, error)

	// Subscription methods
	CreateSubscription(eventID, userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool, isRecurring bool) (string, error)
	UpdateSubscription(subscriptionID, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) error
	GetSubscriptionByProviderCustomerID(providerCustomerID string) (*Subscription, error)
	GetSubscriptionByUserIDAndProductID(userID, productID string) (*Subscription, error)
	CheckUserProductAccess(userID, productID string) (bool, error)

	// User payment methods
	CreateUserPayment(userID, productID, subscriptionID, paymentID, transactionMetadata string, isSuccess bool) (string, error)
}
