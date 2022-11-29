
GO := go

aigaurdd: dependencies
	$(GO) build -o aigaurdd


dependencies:
	$(GO) mod download

clean:
	rm -rf aigaurdd
