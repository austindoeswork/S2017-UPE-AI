package game

import (
	"errors"
	"fmt"
	"log"
	"time"
)

type Player struct {
	id     int //1 or 2
	x      int
	y      int //of top of bar
	length int //above and below center
}

func (p *Player) Position() (int, int) {
	return p.x, p.y
}
func (p *Player) SetPosition(x, y int) {
	p.x = x
	p.y = y
}
func (p *Player) Length() int {
	return p.length
}

type Ball struct {
	x     int
	y     int
	xvel  int
	yvel  int
	speed int // frames to move one gridspace
}

func NewBall(x, y, speed int) *Ball {
	return &Ball{
		x:     x,
		y:     y,
		xvel:  1,
		yvel:  0,
		speed: speed,
	}
}

func (b *Ball) Position() (int, int) {
	return b.x, b.y
}
func (b *Ball) SetPosition(x, y int) {
	b.x = x
	b.y = y
}
func (b *Ball) Velocity() (int, int) {
	return b.xvel, b.yvel
}
func (b *Ball) SetVelocity(xvel, yvel int) {
	b.xvel = xvel
	b.yvel = yvel
}
func (b *Ball) Speed() int {
	return b.speed
}

// TODO abstract, use an interface
type Pong struct {
	p1input   chan []byte
	p2input   chan []byte
	p1cmd     []byte
	p2cmd     []byte
	output    chan []byte //pushes gamestate at framerate?
	quit      chan bool
	done      chan bool
	isStarted bool

	p1   *Player
	p2   *Player
	ball *Ball

	width  int
	height int
	fps    int
	frame  int64

	winner int
}

func New(width, height, fps int) *Pong {
	return &Pong{
		p1input:   make(chan []byte, 5),
		p2input:   make(chan []byte, 5),
		output:    make(chan []byte),
		quit:      make(chan bool),
		done:      make(chan bool),
		isStarted: false,
		ball:      NewBall(width/2, height/2, 4),
		width:     width,
		height:    height,
		fps:       fps,
		frame:     0,
		winner:    -1,
	}
}

// AddPlayer returns an error or 1 or 2 corresponding to the player added
func (p *Pong) AddPlayer() (int, chan []byte, error) {
	if p.p1 == nil {
		p.p1 = &Player{
			id:     1,
			x:      0,
			y:      p.height / 2,
			length: 2,
		}
		return 1, p.p1input, nil
	} else if p.p2 == nil {
		p.p2 = &Player{
			id:     2,
			x:      p.width - 1,
			y:      p.height / 2,
			length: 2,
		}
		return 2, p.p2input, nil
	} else {
		return -1, nil, errors.New("ERROR: 2 Players already joined")
	}
}

// Start returns an output chan and a done chan? TODO, or nil and an error
func (p *Pong) Start() (chan []byte, error) {
	if p.p1 == nil || p.p2 == nil {
		return nil, errors.New("ERROR: not enough players")
	}
	p.isStarted = true

	frameNS := time.Duration(int(1e9) / p.fps)
	clk := time.NewTicker(frameNS)
	go func() {
		log.Println("GAME STARTED YOO")
		for {
			select {
			case <-p.quit:
				log.Println("GAME ABORTED")
				return
			default:
			}

			select {
			case <-clk.C: //nxt frame
				p.frame++
				p.updateInputs()

				if p.p1cmd != nil {
					log.Println("1", p.p1cmd)
				}
				if p.p2cmd != nil {
					log.Println("2", p.p2cmd)
				}

				p.updateGame()

				p.p1cmd = nil
				p.p2cmd = nil

				select {
				case p.output <- p.stateJSON(): //send output
				default:
				}
			}
		}
	}()
	return p.output, nil
}

func (p *Pong) Quit() {
	log.Println("ABORTING GAME...")
	p.quit <- true
}
func (p *Pong) IsStarted() bool {
	return p.IsStarted()
}
func (p *Pong) Winner() int {
	return p.winner
}

func (p *Pong) updateGame() {
	p1string := string(p.p1cmd)
	p2string := string(p.p2cmd)
	controlPlayer(p.p1, p1string, 0, p.height)
	controlPlayer(p.p2, p2string, 0, p.height)

	p1x, p1y := p.p1.Position()
	p2x, p2y := p.p2.Position()
	p1length := p.p1.Length()
	p2length := p.p2.Length()

	if p.frame%int64(p.ball.speed) == 0 { //ball should move
		x, y := p.ball.Position()
		xvel, yvel := p.ball.Velocity()
		xnext := x + xvel
		ynext := y + yvel

		//player collisions
		if xnext == p1x {
			if ynext >= p1y && ynext < p1y+p1length {
				xvel = xvel * -1
				yvel = -1
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			} else if ynext == p1y+p1length {
				xvel = xvel * -1
				yvel = 0
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			} else if ynext <= p1y+2*p1length && ynext > p1y+p1length {
				xvel = xvel * -1
				yvel = 1
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			}
		}
		if xnext == p2x {
			if ynext >= p2y && ynext < p2y+p2length {
				xvel = xvel * -1
				yvel = -1
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			} else if ynext == p2y+p2length {
				xvel = xvel * -1
				yvel = 0
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			} else if ynext <= p2y+2*p2length && ynext > p2y+p2length {
				xvel = xvel * -1
				yvel = 1
				xnext = x + xvel
				p.ball.SetVelocity(xvel, yvel)
			}
		}

		//wall collisions
		if xnext >= p.width || xnext < 0 {
			xnext = p.width / 2
			ynext = p.height / 2
		}
		if ynext >= p.height || ynext < 0 {
			yvel = yvel * -1
			ynext = y + yvel
			p.ball.SetVelocity(xvel, yvel)
		}
		p.ball.SetPosition(xnext, ynext)
	}
}

func controlPlayer(p *Player, input string, upBound, bottomBound int) {
	dx := 0
	dy := 0
	x, y := p.Position()
	length := p.Length()
	switch input {
	case "up":
		if y <= upBound {
			break
		} else {
			dy = -1
		}
	case "down":
		if y+length >= bottomBound {
			break
		} else {
			dy = 1
		}
	}
	p.SetPosition(x+dx, y+dy)
}

func (p *Pong) updateInputs() {
	p1done := false
	p2done := false
	for !p1done {
		select {
		case cmd := <-p.p1input:
			p.p1cmd = cmd
		default:
			p1done = true
		}
	}
	for !p2done {
		select {
		case cmd := <-p.p2input:
			p.p2cmd = cmd
		default:
			p2done = true
		}
	}
}

func (p *Pong) stateJSON() []byte {
	l := p.p1.Length()*2 + 1
	p1x, p1y := p.p1.Position()
	p2x, p2y := p.p2.Position()
	bx, by := p.ball.Position()
	outString := fmt.Sprintf(`{
	"w": %d,
	"h": %d,
	"p1x":%d,
	"p1y":%d,
	"p2x":%d,
	"p2y":%d,
	"l": %d,
	"bx":%d,
	"by":%d
}`, p.width, p.height, p1x, p1y, p2x, p2y, l, bx, by)
	return []byte(outString)
}
