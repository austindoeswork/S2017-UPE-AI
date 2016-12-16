package pong

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// TODO find a better place for this
const (
	NOTREADY = 0 // not enough players / not setup
	READY    = 1 // ready to start
	RUNNING  = 2 // playing now
	DONE     = 3 // done, clean me up
)

type Player struct {
	id     int //1 or 2
	x      int
	y      int //of top of bar
	length int //above and below center
	score  int
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
func (p *Player) Score() int {
	return p.score
}
func (p *Player) Goal() int {
	p.score += 1
	return p.score
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

type Pong struct {
	p1input chan []byte
	p2input chan []byte
	p1cmd   []byte
	p2cmd   []byte
	output  chan []byte //pushes gamestate at framerate
	quit    chan bool
	status  int

	p1   *Player
	p2   *Player
	ball *Ball

	width  int
	height int
	fps    int
	frame  int64

	winner int
}

func New(width, height, fps int) (*Pong, chan []byte) {
	outputChan := make(chan []byte)
	return &Pong{
		p1input: make(chan []byte, 5),
		p2input: make(chan []byte, 5),
		output:  outputChan,
		quit:    make(chan bool),
		status:  NOTREADY,
		ball:    NewBall(width/2, height/2, 4),
		width:   width,
		height:  height,
		fps:     fps,
		frame:   0,
		winner:  -1,
	}, outputChan
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
		p.status = READY
		return 2, p.p2input, nil
	} else {
		return -1, nil, errors.New("ERROR: 2 Players already joined")
	}
}

// Start returns an output chan and a done chan? TODO, or nil and an error
func (p *Pong) Start() error {
	if p.p1 == nil || p.p2 == nil {
		return errors.New("ERROR: not enough players")
	}
	p.status = RUNNING

	frameNS := time.Duration(int(1e9) / p.fps)
	clk := time.NewTicker(frameNS)
	go func() {
		log.Println("GAME STARTED YOO")
		for {

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
				if p.p1.Score() >= 10 || p.p2.Score() >= 10 {
					p.status = DONE
					log.Println("GAME DIED OF NATURAL CAUSES")
					return
				}
			case <-p.quit:
				log.Println("GAME ABORTED")
				return
			}
		}
	}()
	return nil
}

func (p *Pong) Quit() {
	log.Println("ABORTING GAME...")
	p.status = DONE
	select {
	case p.quit <- true:
	default:
	}
}
func (p *Pong) Status() int {
	return p.status
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
		if xnext >= p.width {
			p1score := p.p1.Goal()
			log.Printf("p1 scored: %d", p1score)
			xnext = p.width / 2
			ynext = p.height / 2
		} else if xnext < 0 {
			p2score := p.p2.Goal()
			log.Printf("p2 scored: %d", p2score)
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
	p1s := p.p1.Score()
	p2s := p.p2.Score()
	bx, by := p.ball.Position()
	outString := fmt.Sprintf(`{
	"type": "state",
	"w": %d,
	"h": %d,
	"p1x":%d,
	"p1y":%d,
	"p1s": %d,
	"p2x":%d,
	"p2y":%d,
	"p2s":%d,
	"l": %d,
	"bx":%d,
	"by":%d
}`, p.width, p.height, p1x, p1y, p1s, p2x, p2y, p2s, l, bx, by)
	return []byte(outString)
}