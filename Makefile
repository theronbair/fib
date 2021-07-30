bin/fib:
	go mod tidy
	go build -o bin/fib src/fib/main.go

clean:
	rm -f bin/fib

.PHONY:	test correct-value correct-count malformed-fib malformed-count purgecheck bad-url k6

test: correct-value correct-count malformed-fib malformed-count purgecheck bad-url k6

correct-value:
	./tests/correct-value

correct-count:
	./tests/correct-count

malformed-fib:
	./tests/malformed-fib

malformed-count:
	./tests/malformed-count

purgecheck:
	./tests/purgecheck

bad-url:
	./tests/bad-url

k6:
	k6 run --vus 100 --duration 10s ./tests/load.js

all: bin/fib
