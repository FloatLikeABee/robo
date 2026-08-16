module github.com/robo/morphgraph-worker

go 1.24.1

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/joho/godotenv v1.5.1
	github.com/robo/morphgraph v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4 // indirect
)

replace github.com/robo/morphgraph => ../pkg/morphgraph
