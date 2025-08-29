package fun

type enum interface {
	Names() []string
}

type displayEnum interface {
	DisplayNames() []string
	Names() []string
}
