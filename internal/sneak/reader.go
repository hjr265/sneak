package sneak

type Reader interface {
	Header() Header
	Read([]byte) (int, error)
}
