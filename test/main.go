package main

func main() {
	f()
}

func f() {}
func g() {}

type A struct{}

func (a *A) Func() {}

/*
 *
 *func h() {}
 *
 *
 *func Func() {}
 *
 *type B struct{}
 *
 *func (b B) B() {}
 */
