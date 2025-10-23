package types

type ListOption struct {
	Page  int  // 当前页
	Limit int  // 页大小
	All   bool // 获取所有
}

func (l ListOption) GetPage() int {
	if l.Page <= 0 {
		l.Page = 1
	}
	return l.Page
}

func (l ListOption) GetLimit() int {
	if l.Limit <= 0 {
		l.Limit = 1000
	}
	return l.Limit
}
