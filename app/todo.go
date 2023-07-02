package app

type singleToDoI interface {
	Has() bool
	IsDone() bool
	HasAndNotDone() bool
	HasAndIsDone() bool
	MakeDone()
}

type singleToDo struct {
	has  bool
	done bool
}

func newSingleToDo() *singleToDo {
	return &singleToDo{
		has:  true,
		done: false,
	}
}

var _ singleToDoI = &singleToDo{}

func (c *singleToDo) Has() bool {
	return c.has
}

func (c *singleToDo) IsDone() bool {
	return c.done
}

func (c *singleToDo) HasAndIsDone() bool {
	return c.Has() && c.IsDone()
}

func (c *singleToDo) HasAndNotDone() bool {
	return c.Has() && !c.IsDone()
}

func (c *singleToDo) MakeDone() {
	c.done = true
}
