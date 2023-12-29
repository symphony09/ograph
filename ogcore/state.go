package ogcore

type State interface {
	Get(key any) (any, bool)
	Set(key any, val any)
	Update(key any, updateFunc func(val any) any)
}
