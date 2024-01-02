package player

import (
	"errors"
	"regexp"
)

// Represents a frag
type Death struct {
	Murderer string
	Victim   string
	Means    int
	Solo     bool // self-frag
}

type ObituraryPattern struct {
	matchstr string
	mod      int
	regex    *regexp.Regexp
}

// All possible means of death
const (
	ModUnknown = iota
	ModBlaster
	ModShotgun
	ModSShotgun
	ModMachinegun
	ModChaingun
	ModGrenade
	ModGSplash
	ModRocket
	ModRSplash
	ModHyperblaster
	ModRailgun
	ModBFGLaser
	ModBFGBlast
	ModBFGEffect
	ModHandgrenade // hit with grenade
	ModHGSplash
	ModWater
	ModSlime
	ModLava
	ModCrush
	ModTelefrag
	ModFalling
	ModSuicide
	ModHeldGrenade
	ModExplosive
	ModBarrel
	ModBomb
	ModExit
	ModSplash
	ModTargetLaser
	ModTriggerHurt
	ModHit
	ModTargetBlaster
	ModFriendlyFire
)

var (
	SelfPatterns = []ObituraryPattern{
		{
			matchstr: "(.+) suicides",
			mod:      ModSuicide,
		},
		{
			matchstr: "(.+) cratered",
			mod:      ModFalling,
		},
		{
			matchstr: "(.+) was squished",
			mod:      ModCrush,
		},
		{
			matchstr: "(.+) sank like a rock",
			mod:      ModWater,
		},
		{
			matchstr: "(.+) melted",
			mod:      ModSlime,
		},
		{
			matchstr: "(.+) does a back flip into the lava",
			mod:      ModLava,
		},
		{
			matchstr: "(.+) blew up",
			mod:      ModExplosive,
		},
		{
			matchstr: "(.+) found a way out",
			mod:      ModExit,
		},
		{
			matchstr: "(.+) saw the light",
			mod:      ModTargetLaser,
		},
		{
			matchstr: "(.+) got blasted",
			mod:      ModTargetBlaster,
		},
		{
			matchstr: "(.+) was in the wrong place",
			mod:      ModSplash,
		},
		{
			matchstr: "(.+) tried to put the pin back in",
			mod:      ModHeldGrenade,
		},
		{
			matchstr: "(.+) tripped on .+ own grenade",
			mod:      ModGSplash,
		},
		{
			matchstr: "(.+) blew .+self up",
			mod:      ModRSplash,
		},
		{
			matchstr: "(.+) should have used a smaller gun",
			mod:      ModBFGBlast,
		},
		{
			matchstr: "(.+) killed .+self",
			mod:      ModSuicide,
		},
		{
			matchstr: "(.+) died",
			mod:      ModUnknown,
		},
	}

	MutualPatterns = []ObituraryPattern{
		{
			matchstr: "(.+) was blasted by (.+)",
			mod:      ModBlaster,
		},
		{
			matchstr: "(.+) was gunned down by (.+)",
			mod:      ModShotgun,
		},
		{
			matchstr: "(.+) was blown away by (.+)'s super shotgun",
			mod:      ModSShotgun,
		},
		{
			matchstr: "(.+) was machinegunned by (.+)",
			mod:      ModMachinegun,
		},
		{
			matchstr: "(.+) was cut in half by (.+)'s chaingun",
			mod:      ModChaingun,
		},
		{
			matchstr: "(.+) was popped by (.+)'s grenade",
			mod:      ModGrenade,
		},
		{
			matchstr: "(.+) was shredded by (.+)'s shrapnel",
			mod:      ModGSplash,
		},
		{
			matchstr: "(.+) ate (.+)'s rocket",
			mod:      ModRocket,
		},
		{
			matchstr: "(.+) almost dodged (.+)'s rocket",
			mod:      ModRSplash,
		},
		{
			matchstr: "(.+) was melted by (.+)'s hyperblaster",
			mod:      ModHyperblaster,
		},
		{
			matchstr: "(.+) was railed by (.+)",
			mod:      ModRailgun,
		},
		{
			matchstr: "(.+) saw the pretty lights from (.+)'s BFG",
			mod:      ModBFGLaser,
		},
		{
			matchstr: "(.+) was disintegrated by (.+)'s BFG blast",
			mod:      ModBFGBlast,
		},
		{
			matchstr: "(.+) couldn't hide from (.+)'s BFG",
			mod:      ModBFGEffect,
		},
		{
			matchstr: "(.+) caught (.+)'s handgrenade",
			mod:      ModHandgrenade,
		},
		{
			matchstr: "(.+) didn't see (.+)'s handgrenade",
			mod:      ModHGSplash,
		},
		{
			matchstr: "(.+) feels (.+)'s pain",
			mod:      ModHeldGrenade,
		},
		{
			matchstr: "(.+) tried to invade (.+)'s personal space",
			mod:      ModTelefrag,
		},
	}
)

// Figure out who killed who and how.
//
// Uses an obituary to figure out the who and how.
func CalculateDeath(obit string) (Death, error) {
	// frags involving 2 people are more common, do them first
	for i, frag := range MutualPatterns {
		death := Death{}
		if frag.regex == nil {
			pattern, err := regexp.Compile(frag.matchstr)
			if err != nil {
				continue
			}
			MutualPatterns[i].regex = pattern
		}

		if MutualPatterns[i].regex.Match([]byte(obit)) {
			submatches := MutualPatterns[i].regex.FindAllStringSubmatch(obit, -1)
			death.Means = frag.mod
			death.Victim = submatches[0][1]
			death.Murderer = submatches[0][2]
			death.Solo = false
			return death, nil
		}
	}

	// frags involving 1 person
	for i, frag := range SelfPatterns {
		death := Death{}
		if frag.regex == nil {
			pattern, err := regexp.Compile(frag.matchstr)
			if err != nil {
				continue
			}
			SelfPatterns[i].regex = pattern
		}

		if SelfPatterns[i].regex.Match([]byte(obit)) {
			submatches := SelfPatterns[i].regex.FindAllStringSubmatch(obit, -1)
			death.Means = frag.mod
			death.Victim = submatches[0][1]
			death.Murderer = ""
			death.Solo = true
			return death, nil
		}
	}

	return Death{}, errors.New("obituary not recognised")
}

// Get a string representation of the means of death index
func MODToString(index int) string {
	switch index {
	case ModBlaster:
		return "blaster"
	case ModShotgun:
		return "shotgun"
	case ModSShotgun:
		return "super shotgun"
	case ModMachinegun:
		return "machinegun"
	case ModChaingun:
		return "chaingun"
	case ModGrenade:
		return "grenade"
	case ModGSplash:
		return "grenade"
	case ModRocket:
		return "rocket"
	case ModRSplash:
		return "rocket"
	case ModHyperblaster:
		return "hyperblaster"
	case ModRailgun:
		return "railgun"
	case ModBFGLaser:
		return "BFG"
	case ModBFGBlast:
		return "BFG"
	case ModBFGEffect:
		return "BFG"
	case ModHandgrenade:
		return "handgrenade"
	case ModHGSplash:
		return "handgrenade"
	case ModWater:
		return "water"
	case ModSlime:
		return "slime"
	case ModLava:
		return "lava"
	case ModCrush:
		return "crush"
	case ModTelefrag:
		return "telefrag"
	case ModFalling:
		return "falling"
	case ModSuicide:
		return "suicide"
	case ModHeldGrenade:
		return "handgrenade"
	case ModExplosive:
		return "explosive"
	case ModBarrel:
		return "explosive"
	case ModBomb:
		return "explosive"
	case ModExit:
		return "exit"
	case ModSplash:
		return "splash"
	case ModTargetLaser:
		return "laser"
	case ModTriggerHurt:
		return "trigger"
	case ModHit:
		return "hit"
	case ModTargetBlaster:
		return "targetblaster"
	case ModFriendlyFire:
		return "friendlyfire"
	}
	return "unknown"
}
