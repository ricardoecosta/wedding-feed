package domain

type Message struct {
	Sender  string
	Message string
	Time    uint64
}

type Error struct {
	Code    int
	Message string
}
