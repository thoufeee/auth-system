package dto

// user register

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=4,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	Age      string `json:"age" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// update user profile >>> request

type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Phone string `json:"phone" binding:"omitempty"`
}

// forgot password >> request

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// reset password >> request

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}
