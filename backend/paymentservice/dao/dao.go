package dao

type DAO interface {
	Infer(dummy string) error
	CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string) (string, error)
	ListProducts() ([]*Product, error)
	CreateUserPurchase(userID string, productID string, transaction_metadata string, is_success bool) (string, error)
	GetRazorpayProductById(razorpayProductID string) (*Product, error)
}
