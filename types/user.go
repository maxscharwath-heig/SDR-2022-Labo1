package types

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}

func FindUser(users []User, id int) *User {
	for _, user := range users {
		if user.Id == id {
			return &user
		}
	}
	return nil
}
