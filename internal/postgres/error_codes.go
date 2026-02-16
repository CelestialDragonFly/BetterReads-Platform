package postgres

// https://www.postgresql.org/docs/11/errcodes-appendix.html
var (
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
)
