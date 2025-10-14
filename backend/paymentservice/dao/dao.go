package dao

type DAO interface {
	Infer(dummy string) error
}
