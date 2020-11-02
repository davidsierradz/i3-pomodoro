i3-pomodoro: *.go go.mod go.sum
	go build -o $@ .

clean:
	rm -f i3-pomodoro

install: i3-pomodoro
	install -D $< ${HOME}/.local/bin/$<

PHONY: clean
