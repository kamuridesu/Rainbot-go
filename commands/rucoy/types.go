package rucoy

import "time"

var (
	rucoyBaseURL     = "https://www.rucoyonline.com"
	rucoyStatsApiURL = "https://rucoystatsapi.net"
)

type RucoyGuildMember struct {
	Name          string
	Level         int
	Online        bool
	CharacterPath string
}

type ParsedRucoyGuildData struct {
	Guild   string
	Members []RucoyGuildMember
}

type RucoyInactiveMember struct {
	Name        string
	DaysOffline int
}

type RucoyGoalMember struct {
	Name    string
	Level   int
	Missing int
}

type RucoyCharacterInfo struct {
	Name       string
	Level      int
	Guild      string
	Title      string
	LastOnline string
}

type RucoyStatsGuildResponse struct {
	Players []RucoyStatsGuildPlayer `json:"players"`
}

type RucoyStatsGuildPlayer struct {
	Name       string `json:"name"`
	LastOnline string `json:"lastOnline"`
}

type RucoyLevelTableEntry struct {
	Level              int   `json:"level"`
	ExpLoss            int64 `json:"expLoss"`
	GoldBlackOneNeeded int64 `json:"goldBlackOneNeeded"`
}

type RucoyTrainingMonster struct {
	Name       string
	Defense    int
	HP         float64
	Powertrain bool
}

type RucoyTrainingResult struct {
	Mode                  string
	Monster               string
	Efficiency            float64
	DurationSeconds       float64
	MinimumDuration       float64
	MaxDamage             int
	MaxCriticalDamage     int
	NextMonster           string
	RequiredStats         int
	StatsNeededFor1Damage int
	BestShortMonster      string
	BestShortEfficiency   float64
	BestShortDuration     float64
}

type RucoyTrainingAlternative struct {
	Attack          int
	Monster         string
	Efficiency      float64
	DurationSeconds float64
}

type RucoyUpskillOptions struct {
	DailyHours   int
	Vocation     string
	ManaPerSkill int64
}

type RucoyUpskillManaEstimate struct {
	TotalMana   int64
	MinPotions  int64
	MaxPotions  int64
	TotalArrows int64
	ArrowCost   int64
	MinCost     int64
	MaxCost     int64
}

const rucoyMinimumTrainingDurationSeconds = 8 * 60
const rucoyAFKProfileDelay = 2500 * time.Millisecond
const rucoyUltimateManaPotionMin = 600
const rucoyUltimateManaPotionMax = 900
const rucoyUltimateManaPotionPackSize = 200
const rucoyUltimateManaPotionPackGold = 130000
const rucoyPallyArrowsPerSkill = 4
const rucoyPallySkillsPerSecond = 1
const rucoyPallyArrowBundleSize = 500
const rucoyPallyArrowBundleGold = 1000
