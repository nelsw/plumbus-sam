package arbo

type Client int

const (
	amg   Client = 155
	cbsi  Client = 132
	inuvo Client = 173
)

func Clients() []Client {
	return []Client{amg, cbsi, inuvo}
}
