// SDR - Labo 1
// Nicolas Crausaz & Maxime Scharwath

package types

// User represents an authenticated user of the application
type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
}
