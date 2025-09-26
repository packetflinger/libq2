package flags

import (
	"slices"
	"strings"
)

const (
	NoHealth       = 1 << 0
	NoItems        = 1 << 1
	WeaponsStay    = 1 << 2
	NoFalling      = 1 << 3
	InstantItems   = 1 << 4
	SameLevel      = 1 << 5
	SkinTeams      = 1 << 6
	ModelTeams     = 1 << 7
	NoFriendlyFire = 1 << 8
	SpawnFarthest  = 1 << 9
	ForceRespawn   = 1 << 10
	NoArmor        = 1 << 11
	AllowExit      = 1 << 12
	InfiniteAmmo   = 1 << 13
	QuadDrop       = 1 << 14
	FixedFOV       = 1 << 15
)

var (
	FreeForAllServer = WeaponsStay | InstantItems | ForceRespawn
	TDMServer        = SpawnFarthest | ForceRespawn | QuadDrop
)

// Modify will return a new dmflag value based on a string of instructions.
// Each instruction should be prefixed with either "+" or "-" for including
// or excluding and separated by spaces. Each is an abbreviation of the
// related flag, e.g. "ia" for "infinite ammo", "qd" for "quad drop", etc.
func Modify(val int, inst string) int {
	out := val
	var add, remove []string
	for _, i := range strings.Fields(inst) {
		if op, found := strings.CutPrefix(i, "+"); found {
			add = append(add, op)
		}
		if op, found := strings.CutPrefix(i, "-"); found {
			remove = append(remove, op)
		}
	}
	for _, a := range add {
		switch a {
		case "nh":
			out |= NoHealth
		case "ni":
			out |= NoItems
		case "ws":
			out |= WeaponsStay
		case "fd":
			out &= ^NoFalling // +fd is wanting damage, so remove no falling damage
		case "ii":
			out |= InstantItems
		case "sl":
			out |= SameLevel
		case "st":
			out |= SkinTeams
		case "mt":
			out |= ModelTeams
		case "ff":
			out &= ^NoFriendlyFire // +ff would remove no friendly fire
		case "sf":
			out |= SpawnFarthest
		case "fr":
			out |= ForceRespawn
		case "na":
			out |= NoArmor
		case "ae":
			out |= AllowExit
		case "ia":
			out |= InfiniteAmmo
		case "qd":
			out |= QuadDrop
		case "fov":
			out |= FixedFOV
		}
	}

	for _, r := range remove {
		switch r {
		case "nh":
			out &= ^NoHealth
		case "ni":
			out &= ^NoItems
		case "ws":
			out &= ^WeaponsStay
		case "fd":
			out |= NoFalling // -fd is not wanting damage, so add no falling damage
		case "ii":
			out &= ^InstantItems
		case "sl":
			out &= ^SameLevel
		case "st":
			out &= ^SkinTeams
		case "mt":
			out &= ^ModelTeams
		case "ff":
			out |= NoFriendlyFire // -ff would add no friendly fire
		case "sf":
			out &= ^SpawnFarthest
		case "fr":
			out &= ^ForceRespawn
		case "na":
			out &= ^NoArmor
		case "ae":
			out &= ^AllowExit
		case "ia":
			out &= ^InfiniteAmmo
		case "qd":
			out &= ^QuadDrop
		case "fov":
			out &= ^FixedFOV
		}
	}
	return out
}

// ToSring will return a string representing the dmflags in english. The flags
// will be comma delimited and sorted alphabetically (ascending).
func ToString(f int) string {
	var flags []string
	if (f & NoHealth) > 0 {
		flags = append(flags, "no health")
	}
	if (f & NoItems) > 0 {
		flags = append(flags, "no items")
	}
	if (f & WeaponsStay) > 0 {
		flags = append(flags, "weapons stay")
	}
	if (f & NoFalling) > 0 {
		flags = append(flags, "no falling damage")
	}
	if (f & InstantItems) > 0 {
		flags = append(flags, "instant items")
	}
	if (f & SameLevel) > 0 {
		flags = append(flags, "same level")
	}
	if (f & SkinTeams) > 0 {
		flags = append(flags, "teams by skin")
	}
	if (f & ModelTeams) > 0 {
		flags = append(flags, "teams by model")
	}
	if (f & NoFriendlyFire) > 0 {
		flags = append(flags, "no friendly fire")
	}
	if (f & NoHealth) > 0 {
		flags = append(flags, "spawn farthest")
	}
	if (f & ForceRespawn) > 0 {
		flags = append(flags, "force respawn")
	}
	if (f & NoArmor) > 0 {
		flags = append(flags, "no armor")
	}
	if (f & AllowExit) > 0 {
		flags = append(flags, "allow exit")
	}
	if (f & InfiniteAmmo) > 0 {
		flags = append(flags, "infinite ammo")
	}
	if (f & QuadDrop) > 0 {
		flags = append(flags, "quad drop")
	}
	if (f & FixedFOV) > 0 {
		flags = append(flags, "fixed FOV")
	}
	slices.Sort(flags)
	return strings.Join(flags, ", ")
}
