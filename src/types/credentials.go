// SDR - Labo 2
// Nicolas Crausaz & Maxime Scharwath

package types

// Credentials represents login information used by a user to authenticate
type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
