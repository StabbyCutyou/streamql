test:
	go test -v

test_race:
	go test -v -race

deps:
	go get -u honnef.co/go/tools/cmd/staticcheck

check:
	staticcheck $$(go list ./... | grep -v /vendor/)

bench:
	go test -v -bench=. 

benchmem:
	go test -v -benchmem -bench=. 

count:
	grep -v "//" spub.go | grep . | wc -l