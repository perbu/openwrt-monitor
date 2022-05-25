
all: pi tui

pi:
	go build -o bin/pimatrix cmd/main.go

tui:
	go build -o bin/tui tui/main.go
	