package dto

type PushSubscribeRequest struct {
	Endpoint string `json:"endpoint" validate:"required"`
	Keys     struct {
		P256dh string `json:"p256dh" validate:"required"`
		Auth   string `json:"auth" validate:"required"`
	} `json:"keys" validate:"required"`
}

type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint" validate:"required"`
}

type VAPIDPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}
