package fun

type Guard interface {
	Guard(serviceName string, methodName string, state map[string]string) *Result[any]
}
