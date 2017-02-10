/*
TOWER DEFINITIONS, all of these inherit from UnitBase (just like troops do)
*/
package tdef

/* import (
 	"log"
) */

// INVINCIBLE CORE IS USED FOR THE DEMOGAME (so that it never ends)
type InvincibleCore struct {
	UnitBase
}

func (u *InvincibleCore) CheckBuyable(income, bits int) bool {
	return false
}
func (u *InvincibleCore) ReceiveDamage(damage int) {}
func (u *InvincibleCore) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *InvincibleCore) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *InvincibleCore) Birth(owner *Player, opponent *Player) {}
func (u *InvincibleCore) Die(owner *Player, opponent *Player)   {} // player will be able to detect game loss itself

func NewInvincibleCore(x, y, owner int) Unit {
	return &InvincibleCore{UnitBase{
		enum:   -2,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 10000,
		maxhp:  10000,
		hp:     10000,
		speed:  20,
		stride: 0,
		reach:  300,
	}}
}

// TODO: CREATE SEPARATE INSTANCE FOR OBJECTIVES, as on death objectives increase territory size
type Core struct {
	UnitBase
}

func (u *Core) CheckBuyable(income, bits int) bool {
	return false
}
func (u *Core) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Core) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Core) Iterate(owner *Player, opponent *Player) {
	if u.target != nil {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *Core) Birth(owner *Player, opponent *Player) {}
func (u *Core) Die(owner *Player, opponent *Player)   {} // player will be able to detect game loss itself

func NewCore(x, y, owner int) Unit {
	return &Core{UnitBase{
		enum:   -2,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 100,
		maxhp:  10000,
		hp:     10000,
		speed:  20,
		stride: 0,
		reach:  300,
	}}
}

func NewObjective(x, y, owner int) Unit {
	return &Core{UnitBase{
		enum:   -1,
		x:      x,
		y:      y,
		owner:  owner,
		damage: 30,
		maxhp:  1000,
		hp:     1000,
		speed:  15,
		stride: 0,
		reach:  150,
	}}
}

/*
(Peashooter) [Aggro]
Quick deploy tower with high DPS, high speed and medium range. Low cost, low HP.
*/
type Peashooter struct {
	UnitBase
}

func (u *Peashooter) CheckBuyable(income, bits int) bool {
	return income >= 100 && bits >= 500
}
func (u *Peashooter) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Peashooter) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Peashooter) Iterate(owner *Player, opponent *Player) {
	if u.target != nil && u.enabled {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *Peashooter) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 500)
	owner.SetIncome(owner.Income() - 100)
}
func (u *Peashooter) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 100)
}

func NewPeashooter(x, y, owner int) Unit {
	return &Peashooter{UnitBase{
		enum:     50,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   10,
		maxhp:    200,
		hp:       200,
		speed:    5,
		stride:   0,
		reach:    300,
		enabled:  true,
		infected: false,
	}}
}

/*
(Firewall) [Midrange]
Medium range low-DPS AOE damage tower with high HP and medium cost.
*/
type Firewall struct {
	UnitBase
}

func (u *Firewall) CheckBuyable(income, bits int) bool {
	return income >= 150 && bits >= 1000
}
func (u *Firewall) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Firewall) Prep(owner *Player, opponent *Player) {}
func (u *Firewall) Iterate(owner *Player, opponent *Player) {
	if u.enabled == true {
		for _, element := range opponent.Units {
			if intAbsDiff(element.X(), u.x) <= u.reach && intAbsDiff(element.Y(), u.y) <= u.reach {
				element.ReceiveDamage(u.damage)
			}
		}
	}
}

func (u *Firewall) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
	owner.SetIncome(owner.Income() - 150)
}
func (u *Firewall) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 150)
}

func NewFirewall(x, y, owner int) Unit {
	return &Firewall{UnitBase{
		enum:     51,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   10,
		maxhp:    3000,
		hp:       3000,
		speed:    5,
		stride:   0,
		reach:    300,
		enabled:  true,
		infected: false,
	}}
}

/*
(Guardian) [Control]
Heavy DPS tower with medium reach, high speed, low HP and high cost.
*/
type Guardian struct {
	UnitBase
}

func (u *Guardian) CheckBuyable(income, bits int) bool {
	return income >= 300 && bits >= 5000
}
func (u *Guardian) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Guardian) Prep(owner *Player, opponent *Player) {
	if !u.VerifyTarget() {
		unit, _ := opponent.FindClosestUnit(u)
		u.target = unit
	}
}
func (u *Guardian) Iterate(owner *Player, opponent *Player) {
	if u.target != nil && u.enabled {
		u.target.ReceiveDamage(u.damage)
	}
}

func (u *Guardian) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 5000)
	owner.SetIncome(owner.Income() - 300)
}
func (u *Guardian) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 300)
}

func NewGuardian(x, y, owner int) Unit {
	return &Guardian{UnitBase{
		enum:     52,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   60,
		maxhp:    600,
		hp:       600,
		speed:    3,
		stride:   0,
		reach:    250,
		enabled:  true,
		infected: false,
	}}
}

/*
(Bank)
Standard bank that increases your income by a standard fixed amount.
You start with three banks in the plots closest to your core, which grants you your base income.
*/
type Bank struct {
	UnitBase
}

func (u *Bank) CheckBuyable(income, bits int) bool {
	return bits >= 1000
}
func (u *Bank) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Bank) Prep(owner *Player, opponent *Player) {}

func (u *Bank) Iterate(owner *Player, opponent *Player) {}

func (u *Bank) SetEnabled(owner *Player, enable bool) { // override of UnitBase's SetEnabled()
	if u.enabled == true && enable == false {
		owner.SetIncome(owner.Income() - 100)
	} else if u.enabled == false && enable == true {
		owner.SetIncome(owner.Income() + 100)
	}
	u.enabled = enable
}

func (u *Bank) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
	owner.SetIncome(owner.Income() + 100)
}
func (u *Bank) Die(owner *Player, opponent *Player) {
	if u.enabled == true { // to avoid corner case where a disabled bank is killed before reenabling
		owner.SetIncome(owner.Income() - 100)
	}
}

func NewBank(x, y, owner int) Unit {
	return &Bank{UnitBase{
		enum:     53,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   10,
		maxhp:    200,
		hp:       200,
		speed:    3,
		stride:   0,
		reach:    300,
		enabled:  true,
		infected: false,
	}}
}

/*
(Junkyard) [Aggro]
When units die within a certain range of the Junkyard, the owner of the Junkyard will make a fixed, decent chunk of change back.
The idea is to strategically place Junkyards deep into the fray, as they are cheap and somewhat fragile anyway.
Junkyards are not a very stable source of income, and are completely unviable in the mid-late game.
*/
type Junkyard struct {
	UnitBase
}

func (u *Junkyard) CheckBuyable(income, bits int) bool {
	return bits >= 1000
}
func (u *Junkyard) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Junkyard) Prep(owner *Player, opponent *Player) {}

func (u *Junkyard) Iterate(owner *Player, opponent *Player) {}

func (u *Junkyard) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
}

func (u *Junkyard) Die(owner *Player, opponent *Player) {}

func NewJunkyard(x, y, owner int) Unit {
	return &Junkyard{UnitBase{
		enum:     54,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   10,
		maxhp:    600,
		hp:       600,
		speed:    3,
		stride:   0,
		reach:    500,
		enabled:  true,
		infected: false,
	}}
}

/*
(Start up) [Midrange]
Bank that increases your income by a small fixed amount that grows linearly over time up till a max value.
Takes time to become economically viable, and then becomes less effective than other options after a certain period of time.
For a decently hefty price, can be transformed into a Corporation.
*/
type StartUp struct {
	UnitBase
	income int
}

func (u *StartUp) CheckBuyable(income, bits int) bool {
	return bits >= 1000
}
func (u *StartUp) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *StartUp) Prep(owner *Player, opponent *Player) {}

func (u *StartUp) Iterate(owner *Player, opponent *Player) {
	if u.income < 500 && u.enabled == true {
		u.income += 4
		owner.SetIncome(owner.Income() + 4)
	}
}

func (u *StartUp) SetEnabled(owner *Player, enable bool) { // override of UnitBase's SetEnabled()
	if u.enabled == true && enable == false {
		owner.SetIncome(owner.Income() - u.income)
	} else if u.enabled == false && enable == true {
		owner.SetIncome(owner.Income() + u.income)
	}
	u.enabled = enable
}

func (u *StartUp) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
}
func (u *StartUp) Die(owner *Player, opponent *Player) {
	if u.enabled == true { // to avoid corner case where a disabled bank is killed before reenabling
		owner.SetIncome(owner.Income() - u.income)
	}
}

func NewStartUp(x, y, owner int) Unit {
	return &StartUp{
		UnitBase: UnitBase{
			enum:     55,
			x:        x,
			y:        y,
			owner:    owner,
			damage:   10,
			maxhp:    200,
			hp:       200,
			speed:    3,
			stride:   0,
			reach:    300,
			enabled:  true,
			infected: false,
		},
		income: 100,
	}
}

/*
(Corporation) [Control]
Expensive bank that increases your income by a fixed amount that is multiplied exponentially by the number of how many of these towers you own.
Meant to snowball money by a crazy amount in the late game.
*/
type Corporation struct {
	UnitBase
	income int
}

func (u *Corporation) CheckBuyable(income, bits int) bool {
	return bits >= 25000
}
func (u *Corporation) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Corporation) Prep(owner *Player, opponent *Player) {}

func (u *Corporation) Iterate(owner *Player, opponent *Player) {
	if u.enabled == false {
		return
	}
	corpCount := 0
	for _, element := range owner.Towers {
		if element == nil {
			continue
		} else if element.Enum() == 56 {
			corpCount++
		}
	}
	newValue := corpCount * corpCount * 200
	if u.income != newValue {
		owner.SetIncome(owner.Income() - u.income + newValue)
		u.income = newValue
	}
}

func (u *Corporation) SetEnabled(owner *Player, enable bool) { // override of UnitBase's SetEnabled()
	if u.enabled == true && enable == false {
		owner.SetIncome(owner.Income() - u.income)
	} else if u.enabled == false && enable == true {
		owner.SetIncome(owner.Income() + u.income)
	}
	u.enabled = enable
}

func (u *Corporation) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 25000)
}
func (u *Corporation) Die(owner *Player, opponent *Player) {
	if u.enabled == true { // to avoid corner case where a disabled bank is killed before reenabling
		owner.SetIncome(owner.Income() - u.income)
	}
}

func NewCorporation(x, y, owner int) Unit {
	return &Corporation{
		UnitBase: UnitBase{
			enum:     56,
			x:        x,
			y:        y,
			owner:    owner,
			damage:   10,
			maxhp:    200,
			hp:       200,
			speed:    3,
			stride:   0,
			reach:    300,
			enabled:  true,
			infected: false,
		},
		income: 0,
	}
}

/*
(Warp Drive) [Aggro]
Friendly troops in the Warp Drive’s lane spawn from the warp drive instead of from the original core. Decently cheap, but are fragile targets.
*/
type WarpDrive struct {
	UnitBase
}

func (u *WarpDrive) CheckBuyable(income, bits int) bool {
	return bits >= 1000
}
func (u *WarpDrive) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *WarpDrive) Prep(owner *Player, opponent *Player) {}

func (u *WarpDrive) Iterate(owner *Player, opponent *Player) {}

func (u *WarpDrive) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 1000)
}

func (u *WarpDrive) Die(owner *Player, opponent *Player) {}

func NewWarpDrive(x, y, owner int) Unit {
	return &WarpDrive{UnitBase{
		enum:     57,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   0,
		maxhp:    400,
		hp:       400,
		speed:    1,
		stride:   0,
		reach:    0,
		enabled:  true,
		infected: false,
	}}
}

/*
(EMP Station/Jamming Station) [Midrange]
Crippling tower that slows the stride and speed of enemy troops in a medium sized AOE around it. Pretty beefy towers, with a moderate cost.
*/

type JammingStation struct {
	UnitBase
}

func (u *JammingStation) CheckBuyable(income, bits int) bool {
	return income >= 150 && bits >= 2500
}
func (u *JammingStation) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *JammingStation) Prep(owner *Player, opponent *Player) {}
func (u *JammingStation) Iterate(owner *Player, opponent *Player) {
	if u.enabled == true {
		for _, element := range opponent.Units {
			if intAbsDiff(element.X(), u.x) <= u.reach && intAbsDiff(element.Y(), u.y) <= u.reach {
				if element.Speed() < 15 { // minimum speed
					element.SetSpeed(element.Speed() + 1)
				}
				if element.Stride() > 3 { // minimum stride
					element.SetStride(element.Stride() - 1)
				}
			}
		}
	}
}

func (u *JammingStation) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 2500)
	owner.SetIncome(owner.Income() - 150)
}
func (u *JammingStation) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 150)
}

func NewJammingStation(x, y, owner int) Unit {
	return &JammingStation{UnitBase{
		enum:     58,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   0,
		maxhp:    3000,
		hp:       3000,
		speed:    30,
		stride:   0,
		reach:    200,
		enabled:  true,
		infected: false,
	}}
}

/*
(Hotspot) [Control]
Friendly units in the lane controlled by the Hotspot have their damage increased by a fixed amount, and are healed every iteration by a small amount.
Hotspots are quite expensive and are fragile targets.
*/

type Hotspot struct {
	UnitBase
}

func (u *Hotspot) CheckBuyable(income, bits int) bool {
	return income >= 500 && bits >= 10000
}
func (u *Hotspot) ReceiveDamage(damage int) {
	u.hp -= damage
}
func (u *Hotspot) Prep(owner *Player, opponent *Player) {}
func (u *Hotspot) Iterate(owner *Player, opponent *Player) {
	if u.enabled == true {
		for _, element := range owner.Units {
			if intAbsDiff(element.Y(), u.y) <= 100 {
				element.SetDamage(element.Damage() + 1)
			}
		}
	}
}

func (u *Hotspot) Birth(owner *Player, opponent *Player) {
	owner.SetBits(owner.Bits() - 10000)
	owner.SetIncome(owner.Income() - 500)
}
func (u *Hotspot) Die(owner *Player, opponent *Player) {
	owner.SetIncome(owner.Income() + 500)
}

func NewHotspot(x, y, owner int) Unit {
	return &Hotspot{UnitBase{
		enum:     59,
		x:        x,
		y:        y,
		owner:    owner,
		damage:   0,
		maxhp:    1000,
		hp:       1000,
		speed:    10,
		stride:   0,
		reach:    300,
		enabled:  true,
		infected: false,
	}}
}
