package storage

type Storage interface {
	GetPackSizes() []int
	SetPackSizes(sizes []int) error
}
