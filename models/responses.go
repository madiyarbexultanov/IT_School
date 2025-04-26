package models

// Стандартные модели ответов
type (
    // LoginResponse - ответ на успешный логин
    LoginResponse struct {
        Token   string `json:"token" example:"eyJhbGciOi..."`
        Role    string `json:"role" example:"admin"`
        Expires int64  `json:"expires" example:"1672531200"`
    }

    // MessageResponse - универсальный ответ с сообщением
    MessageResponse struct {
        Message string `json:"message" example:"success message"`
    }

    // ErrorResponse - стандартный формат ошибки
    ErrorResponse struct {
        Error string `json:"error" example:"error description"`
    }

	TokenResponse struct {
		Token   string `json:"token" example:"eyJhbGciOi..."`
		Expires int64  `json:"expires" example:"1672531200"`
	}
)