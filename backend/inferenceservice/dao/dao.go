package dao

type InferenceDAO interface {
	Infer(dummy string) error
}
