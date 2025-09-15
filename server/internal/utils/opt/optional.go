package opt

type Val[T any] struct {
	set   bool
	value T
}

func Some[T any](v T) Val[T] {
	return Val[T]{set: true, value: v}
}

func None[T any]() Val[T] {
	return Val[T]{set: false}
}

func FromRef[T any](ref *T) Val[T] {
	if ref == nil {
		return None[T]()
	}
	return Some(*ref)
}

func (o Val[T]) IsSet() bool {
	return o.set
}

func (o Val[T]) Value() (T, bool) {
	return o.value, o.set
}

func (o Val[T]) MustValue() T {
	if !o.set {
		panic("read from unset opt.Val")
	}
	return o.value
}
