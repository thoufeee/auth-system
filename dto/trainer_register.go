package dto

// trainer register
type TrainerRegisterRequest struct {
	Name            string   `json:"name" binding:"required,min=2,max=100"`
	Email           string   `json:"email" binding:"required,email"`
	Phone           string   `json:"phone" binding:"required"`
	Age             int      `json:"age" binding:"required,min=18,max=100"`
	Password        string   `json:"password" binding:"required,min=8"`
	Bio             string   `json:"bio" binding:"required"`
	Specializations []string `json:"specializations" binding:"required,min=1"`
	ExperienceYears int      `json:"experience_years" binding:"min=0"`
	Certifications  []string `json:"certifications"`
	HourlyRate      float64  `json:"hourly_rate" binding:"gte=0"`
}
