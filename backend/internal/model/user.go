package model

type AuthUser struct {
	ID         string `json:"id"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Patronymic string `json:"patronymic"`
}

type User struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Patronymic string  `json:"patronymic"`
	Phone      string  `json:"phone"`
	Role       string  `json:"role"`
	Position   string  `json:"position"`
	BirthDate  *string `json:"birthDate,omitempty"`
	IsActive   bool    `json:"isActive"`
	CreatedAt  int64   `json:"createdAt"`
	UpdatedAt  int64   `json:"updatedAt"`
}

type CreateUserInput struct {
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Patronymic string  `json:"patronymic"`
	Phone      string  `json:"phone"`
	Password   string  `json:"password"`
	Role       string  `json:"role"`
	Position   string  `json:"position"`
	BirthDate  *string `json:"birthDate"`
}

type UpdateUserInput struct {
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Patronymic string  `json:"patronymic"`
	Phone      string  `json:"phone"`
	Password   string  `json:"password"`
	Role       string  `json:"role"`
	Position   string  `json:"position"`
	BirthDate  *string `json:"birthDate"`
}
