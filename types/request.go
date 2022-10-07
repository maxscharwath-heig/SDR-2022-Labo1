package types

type Request[T any] struct {
	Credentials Credentials `json:"credentials"`
	Data        T           `json:"data"`
}
