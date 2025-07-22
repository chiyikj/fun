package fun

func MockRequest[T any](clientInfo ClientInfo, requestInfo RequestInfo[any]) Result[T] {
	port := randomPort()
	go func() {
		Start(port)
	}()
	c := client(clientInfo.Id, port)
	defer c.Close()
	err := c.WriteJSON(requestInfo)
	if err != nil {
		panic(err)
	}
	result := Result[T]{}
	err = c.ReadJSON(&result)
	if err != nil {
		panic(err)
	}
	return result
}
