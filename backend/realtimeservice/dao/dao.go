package dao

type DAO interface {
	Infer(dummy string) error
	CreateAudioChat(userID string, modelName string, startTime string, endTime string) (string, error)
	UpdateAudioChat(userID string, id string, endTime string) error
}
