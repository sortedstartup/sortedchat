package dao

type Product struct {
	ID          string `db:"id"`
	UserID      string `db:"user_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Price       string `db:"price"`
	Currency    string `db:"currency"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
}
