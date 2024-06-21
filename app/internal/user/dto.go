package user

type CreateUserDTO struct {
	Name             string `json:"name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

type UpdateUserDTO struct {
	UUID                string `json:"uuid,omitempty"`
	Name                string `json:"name,omitempty"`
	Email               string `json:"email,omitempty"`
	OldPassword         string `json:"old_password,omitempty"`
	NewPassword         string `json:"new_password,omitempty"`
	RepeatedNewPassword string `json:"repeated_new_password,omitempty"`
}

type EmailAndPasswordDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
