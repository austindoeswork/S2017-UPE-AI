package tdef

type Controller struct {
	player int
	input  chan<- []byte
	output <-chan []byte
}

func (c *Controller) Player() int {
	return c.player
}

func (c *Controller) Input() chan<- []byte {
	return c.input
}

func (c *Controller) Output() <-chan []byte {
	return c.output
}
