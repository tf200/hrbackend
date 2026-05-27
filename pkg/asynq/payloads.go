package asynq

type EmailDeliveryPayload struct {
	To           string `json:"to"`
	Name         string `json:"name"`
	UserEmail    string `json:"user_email"`
	UserPassword string `json:"user_password"`
}
