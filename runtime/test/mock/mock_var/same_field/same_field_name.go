package same_field

var DB int = 10

type Other struct {
	DB string
}

func Run() *Other {
	db := consume(&Other{
		DB: "test",
	})
	_ = db.DB
	return db
}

func consume(db *Other) *Other {
	return db
}
