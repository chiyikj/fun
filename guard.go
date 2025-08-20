package fun

type Guard interface {
	Guard(ctx Ctx) *Result[any]
}
