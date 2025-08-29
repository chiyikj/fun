package fun

type Enum interface {
	Names() []string
}

type DisplayEnum interface {
	DisplayNames() []string
	Names() []string
}
