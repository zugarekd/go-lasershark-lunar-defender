package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var lander *Lunar

var flame *Flame

var gameObjects = make(map[string]GameObject)

func main() {
	http.HandleFunc("/event", event)
	go http.ListenAndServe(":8090", nil)

	tmpFlame := NewFlame()
	flame = &tmpFlame
	flame.Active = true

	ground := NewGround()
	ground.Active = true

	tmpLander := NewLunar()
	lander = &tmpLander
	lander.SetPosition(600, 600)
	lander.Active = true
	lander.AccelerationX = 0
	lander.AccelerationY = 0

	target := NewTarget()
	target.SetPosition(1000, 2000)
	target.Active = true

	gameObjects["lander"] = lander
	gameObjects["flame"] = flame
	gameObjects["ground"] = &ground
	gameObjects["target"] = &target

	fmt.Printf("r=2000\n")
	fmt.Printf("e=1\n")

	go func() {
		for true {
			updateState()
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for true {
		draw()
	}
}

type Position struct {
	X float64
	Y float64
}

type Line struct {
	X1 float64
	Y1 float64
	X2 float64
	Y2 float64
}

type GameObject interface {
	GetLines() []Line
	GetPosition() Position
	GetActive() bool
	//GetAngle() float64
	//SetAngle(float64)
}

func draw() {
	for _, object := range gameObjects {
		if object.GetActive() {
			position := object.GetPosition()
			command := fmt.Sprintf("s=%v,%v,%v,%v,%v,%v\n", int(object.GetLines()[0].X1+position.X), int(object.GetLines()[0].Y1+position.Y), "0", "0", "0", "0")
			for _, line := range object.GetLines() {
				command = command + fmt.Sprintf("s=%v,%v,%v,%v,%v,%v\n", int(line.X1+position.X), int(line.Y1+position.Y), "4095", "4095", "1", "1")
				command = command + fmt.Sprintf("s=%v,%v,%v,%v,%v,%v\n", int(line.X2+position.X), int(line.Y2+position.Y), "4095", "4095", "1", "1")
			}
			fmt.Printf(command)
		}
	}
}

func event(w http.ResponseWriter, req *http.Request) {
	event, _ := req.URL.Query()["event"]

	key, _ := req.URL.Query()["key"]

	if event[0] == "down" && key[0] == "68" {
		turningLeft = true
	}
	if event[0] == "down" && key[0] == "65" {
		turningRight = true
	}
	if event[0] == "down" && key[0] == "87" {
		thrust = true
	}
	if event[0] == "up" && key[0] == "68" {
		turningLeft = false
	}
	if event[0] == "up" && key[0] == "65" {
		turningRight = false
	}
	if event[0] == "up" && key[0] == "87" {
		thrust = false
	}
	if event[0] == "down" && key[0] == "32" {
		bullet := NewBullet()
		bullet.Angle = lander.Angle
		bullet.Position = lander.Position
		bullet.Active = true

		t := time.Now()
		gameObjects[t.String()] = &bullet
	}
}

var turningRight = false

var turningLeft = false

var thrust = false

func updateState() {
	target, ok := gameObjects["target"]
	if ok {
		t := target.(*Target)
		t.AccelerationY = t.AccelerationY - .5
		t.Position.Y = t.Position.Y + t.AccelerationY
		if t.Position.Y < 100 {
			t.Position.Y = 100
		}
	} else {
		t := NewTarget()
		t.Active = true

		t.SetPosition((rand.Float64()*2000)+100, 2000)
		gameObjects["target"] = &t
	}

	if turningLeft {
		lander.Angle = lander.Angle - 10
	}
	if turningRight {
		lander.Angle = lander.Angle + 10
	}
	if lander.Angle < 0 {
		lander.Angle = lander.Angle + 360
	}
	if lander.Angle > 360 {
		lander.Angle = lander.Angle - 360
	}
	flame.Angle = lander.Angle

	if thrust {
		flame.Active = true

		val := math.Cos(lander.Angle * math.Pi / 180)
		lander.AccelerationY = lander.AccelerationY + (val * THRUST)

		val = math.Sin(lander.Angle * math.Pi / 180)
		lander.AccelerationX = lander.AccelerationX + (val * THRUST * -1)
	} else {
		flame.Active = false
		lander.AccelerationY = lander.AccelerationY + GRAVITY
	}
	landerY := lander.Position.Y + lander.AccelerationY
	landerX := lander.Position.X + lander.AccelerationX

	if landerY < 100 {
		landerY = 100
		if lander.AccelerationY < 0 {
			lander.AccelerationY = 0
		}
		lander.AccelerationX = 0
	}
	if landerY > 2100 {
		landerY = 2100
	}

	lander.SetPosition(landerX, landerY)

	for key, object := range gameObjects {
		if object.GetActive() {
			if bullet, ok := object.(*Bullet); ok {
				x1 := bullet.Position.X - math.Sin(bullet.Angle*math.Pi/180)*60
				y1 := bullet.Position.Y + math.Cos(bullet.Angle*math.Pi/180)*60

				if x1 < 100 || x1 > 2000 || y1 < 100 || y1 > 2000 {
					bullet.Active = false
					delete(gameObjects, key)
				} else {
					bullet.Position.X = x1
					bullet.Position.Y = y1
				}
			}
			if target, ok := object.(*Target); ok {
				for colKey, collision := range gameObjects {
					if col, ok := collision.(*Bullet); ok {
						if target.GetLines()[0].X1+target.Position.X <= col.Lines[0].X1+col.Position.X &&
							col.Lines[0].X1+col.Position.X <= target.GetLines()[2].X1+target.Position.X &&
							target.GetLines()[0].Y1+target.Position.Y <= col.Lines[0].Y1+col.Position.Y &&
							col.Lines[0].Y1+col.Position.Y <= target.GetLines()[2].Y1+target.Position.X {
							delete(gameObjects, colKey)
							delete(gameObjects, "target")
						}
						if target.GetLines()[0].X1+target.Position.X <= col.Lines[0].X2+col.Position.X &&
							col.Lines[0].X2+col.Position.X <= target.GetLines()[2].X1+target.Position.X &&
							target.GetLines()[0].Y1+target.Position.Y <= col.Lines[0].Y2+col.Position.Y &&
							col.Lines[0].Y2+col.Position.Y <= target.GetLines()[2].Y1+target.Position.X {
							delete(gameObjects, colKey)
							delete(gameObjects, "target")
						}
					}
				}
			}
		}
	}
}

//var MAX_SPEED float64  = 250

//var DRAG float64  = 0

var THRUST float64 = .75

var GRAVITY float64 = -.5

type Ground struct {
	Position Position
	Lines    []Line
	Center   Position
	Angle    float64
	Active   bool
}

func NewGround() Ground {
	ground := Ground{
		Center: Position{
			X: 0,
			Y: 0,
		},
		Angle:    0,
		Position: Position{},
		Lines: []Line{{
			X1: 0,
			Y1: 93,
			X2: 2000,
			Y2: 93,
		},
		},
	}
	return ground
}

func (ground Ground) GetLines() []Line {
	return ground.Lines
}

func (ground *Ground) GetPosition() Position {
	return ground.Position
}

func (ground *Ground) GetActive() bool {
	return ground.Active
}

func (ground *Ground) SetPosition(x float64, y float64) {
	ground.Position.X = x
	ground.Position.Y = y
}

type Lunar struct {
	Position      Position
	Lines         []Line
	Center        Position
	Angle         float64
	Active        bool
	AccelerationX float64
	AccelerationY float64
}

func NewLunar() Lunar {
	lunar := Lunar{
		Center: Position{
			X: 40,
			Y: 50,
		},
		Angle:    0,
		Position: Position{},
		Lines: []Line{{
			X1: 0,
			Y1: 0,
			X2: 40,
			Y2: 100,
		}, {
			X1: 40,
			Y1: 100,
			X2: 80,
			Y2: 0,
		}, {
			X1: 80,
			Y1: 0,
			X2: 0,
			Y2: 0,
		},
		},
	}
	return lunar
}

func (lunar Lunar) GetLines() []Line {
	var rotatedLines []Line

	for _, line := range lunar.Lines {
		s := math.Sin(lunar.Angle * (math.Pi / 180))
		c := math.Cos(lunar.Angle * (math.Pi / 180))

		line.X1 = line.X1 - lunar.Center.X
		line.Y1 = line.Y1 - lunar.Center.Y
		line.X2 = line.X2 - lunar.Center.X
		line.Y2 = line.Y2 - lunar.Center.Y

		x1new := line.X1*c - line.Y1*s
		y1new := line.X1*s + line.Y1*c
		x2new := line.X2*c - line.Y2*s
		y2new := line.X2*s + line.Y2*c

		line.X1 = x1new + lunar.Center.X
		line.Y1 = y1new + lunar.Center.Y
		line.X2 = x2new + lunar.Center.X
		line.Y2 = y2new + lunar.Center.Y

		rotatedLines = append(rotatedLines, line)
	}
	return rotatedLines
}

func (lunar *Lunar) GetPosition() Position {
	return lunar.Position
}

func (lunar *Lunar) GetActive() bool {
	return lunar.Active
}

func (lunar *Lunar) SetPosition(x float64, y float64) {
	lunar.Position.X = x
	lunar.Position.Y = y
	flame.SetPosition(x, y)
}

type Flame struct {
	Position Position
	Lines    []Line
	Center   Position
	Angle    float64
	Active   bool
}

func NewFlame() Flame {
	flame := Flame{
		Center: Position{
			X: 40,
			Y: 50,
		},
		Angle:    0,
		Position: Position{},
		Lines: []Line{{
			X1: 20,
			Y1: 0,
			X2: 40,
			Y2: -40,
		}, {
			X1: 40,
			Y1: -40,
			X2: 60,
			Y2: 0,
		},
		},
	}
	return flame
}

func (flame Flame) GetLines() []Line {
	var rotatedLines []Line

	for _, line := range flame.Lines {
		s := math.Sin(flame.Angle * (math.Pi / 180))
		c := math.Cos(flame.Angle * (math.Pi / 180))

		line.X1 = line.X1 - flame.Center.X
		line.Y1 = line.Y1 - flame.Center.Y
		line.X2 = line.X2 - flame.Center.X
		line.Y2 = line.Y2 - flame.Center.Y

		x1new := line.X1*c - line.Y1*s
		y1new := line.X1*s + line.Y1*c
		x2new := line.X2*c - line.Y2*s
		y2new := line.X2*s + line.Y2*c

		line.X1 = x1new + flame.Center.X
		line.Y1 = y1new + flame.Center.Y
		line.X2 = x2new + flame.Center.X
		line.Y2 = y2new + flame.Center.Y

		rotatedLines = append(rotatedLines, line)
	}
	return rotatedLines
}

func (flame *Flame) GetPosition() Position {
	return flame.Position
}

func (flame *Flame) GetActive() bool {
	return flame.Active
}

func (flame *Flame) SetPosition(x float64, y float64) {
	flame.Position.X = x
	flame.Position.Y = y
}

type Bullet struct {
	Position Position
	Lines    []Line
	Center   Position
	Angle    float64
	Active   bool
}

func NewBullet() Bullet {
	bullet := Bullet{
		Center: Position{
			X: 40,
			Y: 50,
		},
		Angle:    0,
		Position: Position{},
		Lines: []Line{{
			X1: 40,
			Y1: 100,
			X2: 40,
			Y2: 150,
		},
		},
	}
	return bullet
}

func (bullet Bullet) GetLines() []Line {
	var rotatedLines []Line

	for _, line := range bullet.Lines {
		s := math.Sin(bullet.Angle * (math.Pi / 180))
		c := math.Cos(bullet.Angle * (math.Pi / 180))

		line.X1 = line.X1 - bullet.Center.X
		line.Y1 = line.Y1 - bullet.Center.Y
		line.X2 = line.X2 - bullet.Center.X
		line.Y2 = line.Y2 - bullet.Center.Y

		x1new := line.X1*c - line.Y1*s
		y1new := line.X1*s + line.Y1*c
		x2new := line.X2*c - line.Y2*s
		y2new := line.X2*s + line.Y2*c

		line.X1 = x1new + bullet.Center.X
		line.Y1 = y1new + bullet.Center.Y
		line.X2 = x2new + bullet.Center.X
		line.Y2 = y2new + bullet.Center.Y

		rotatedLines = append(rotatedLines, line)
	}
	return rotatedLines
}

func (bullet *Bullet) GetPosition() Position {
	return bullet.Position
}

func (bullet *Bullet) GetActive() bool {
	return bullet.Active
}

func (bullet *Bullet) SetPosition(x float64, y float64) {
	bullet.Position.X = x
	bullet.Position.Y = y
}

type Target struct {
	Position      Position
	Lines         []Line
	Center        Position
	Angle         float64
	Active        bool
	AccelerationY float64
}

func NewTarget() Target {
	target := Target{
		Center: Position{
			X: 40,
			Y: 50,
		},
		Angle:    0,
		Position: Position{},
		Lines: []Line{{
			X1: 0,
			Y1: 0,
			X2: 0,
			Y2: 200,
		},
			{
				X1: 0,
				Y1: 200,
				X2: 200,
				Y2: 200,
			},
			{
				X1: 200,
				Y1: 200,
				X2: 200,
				Y2: 0,
			},
			{
				X1: 200,
				Y1: 0,
				X2: 0,
				Y2: 0,
			},
		},
	}
	return target
}

func (target Target) GetLines() []Line {
	var rotatedLines []Line

	for _, line := range target.Lines {
		s := math.Sin(target.Angle * (math.Pi / 180))
		c := math.Cos(target.Angle * (math.Pi / 180))

		line.X1 = line.X1 - target.Center.X
		line.Y1 = line.Y1 - target.Center.Y
		line.X2 = line.X2 - target.Center.X
		line.Y2 = line.Y2 - target.Center.Y

		x1new := line.X1*c - line.Y1*s
		y1new := line.X1*s + line.Y1*c
		x2new := line.X2*c - line.Y2*s
		y2new := line.X2*s + line.Y2*c

		line.X1 = x1new + target.Center.X
		line.Y1 = y1new + target.Center.Y
		line.X2 = x2new + target.Center.X
		line.Y2 = y2new + target.Center.Y

		rotatedLines = append(rotatedLines, line)
	}
	return rotatedLines
}

func (target *Target) GetPosition() Position {
	return target.Position
}

func (target *Target) GetActive() bool {
	return target.Active
}

func (target *Target) SetPosition(x float64, y float64) {
	target.Position.X = x
	target.Position.Y = y
}
