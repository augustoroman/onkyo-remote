package eiscp

type Message string

func (m Message) Code() string {
	return string(m[2:5])
}
func (m Message) Value() string {
	return string(m[5:])
}
