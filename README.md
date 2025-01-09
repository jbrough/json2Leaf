# json2Leaf

json structure > sql schema that can be traversed with postgresql recursive tree query

`go run cmd/schema/main.go ./my/reporting/dir`
`psql -U postgres -d mydb -f output.sql`
