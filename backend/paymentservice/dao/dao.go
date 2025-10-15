package dao

type DAO interface {
	Infer(dummy string) error
	CreateProduct(id string, userID string, name string, description string, amountInCents int64, currency string) (string, error)
	ListProducts() ([]*Product, error)
	CreateUserPurchase(userID string, productID string, transaction_metadata string, is_success bool) (string, error)
}
