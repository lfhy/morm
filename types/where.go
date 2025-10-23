package types

type WhereMode int

const (
	WhereIs WhereMode = iota
	WhereNot
	WhereGt
	WhereLt
	WhereOr
	OrderAsc
	OrderDesc
	WhereGte
	WhereLte
	WhereLike
)
