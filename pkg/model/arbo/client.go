package arbo

type Client int

const (
	cbsi  Client = 132
	inuvo Client = 173
)

func Clients() []Client {
	return []Client{inuvo, cbsi}
}
