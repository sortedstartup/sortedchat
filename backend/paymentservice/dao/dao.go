package dao

type DAO interface {
	Infer(dummy string) error
	CreateProduct(id string, userID string, name string, description string, cost string, currency string) (string, error)
	ListProducts(userID string) ([]*Product, error)
}
