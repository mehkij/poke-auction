package utils

func SafeCloseChannel(ch chan bool) {
	defer func() {
		recover()
	}()
	select {
	case ch <- true:
	default:
	}
	close(ch)
}
