package rucoy

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func RucoyTrain(m *messages.Message) {
	args := *m.Args

	attack, err := strconv.Atoi(args[0])
	if err != nil {
		m.Reply("arma precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	baseLevel, err := strconv.Atoi(args[1])
	if err != nil {
		m.Reply("level precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	statLevel, err := strconv.Atoi(args[2])
	if err != nil {
		m.Reply("skill precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	extraStats, err := strconv.Atoi(args[3])
	if err != nil {
		m.Reply("add precisa ser um numero. Exemplo: /train 5 351 391 -50", emojis.Fail)
		return
	}

	targetEfficiency := 90.0
	if len(args) >= 5 {
		targetEfficiency, err = strconv.ParseFloat(strings.ReplaceAll(args[4], ",", "."), 64)
		if err != nil {
			m.Reply("eficiencia precisa ser um numero. Exemplo: /train 5 351 391 -50 90", emojis.Fail)
			return
		}
	}

	if err := validateRucoyTrainInput(attack, baseLevel, statLevel, extraStats, targetEfficiency); err != nil {
		m.Reply(err.Error(), emojis.Fail)
		return
	}

	afkResult := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, false, targetEfficiency)
	powerResult := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, true, targetEfficiency)

	sb := strings.Builder{}
	fmt.Fprintf(&sb, "Calculadora Train Rucoy\n\n")
	fmt.Fprintf(&sb, "Setup: arma %d | level %d | skill %d | add %+d\n", attack, baseLevel, statLevel, extraStats)
	fmt.Fprintf(&sb, "Skill efetiva: %d\n", statLevel+extraStats)
	fmt.Fprintf(&sb, "Eficiencia alvo: %.0f%%+\n\n", targetEfficiency)
	writeRucoyTrainingResult(&sb, afkResult, targetEfficiency)
	if afkResult.Monster == "" {
		writeRucoyTrainingWeaponAlternatives(&sb, attack, baseLevel, statLevel, extraStats, targetEfficiency, false)
	}
	sb.WriteString("\n")
	writeRucoyTrainingResult(&sb, powerResult, targetEfficiency)
	if powerResult.Monster == "" {
		writeRucoyTrainingWeaponAlternatives(&sb, attack, baseLevel, statLevel, extraStats, targetEfficiency, true)
	}

	m.Reply(sb.String(), emojis.Success)
}

func validateRucoyTrainInput(attack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64) error {
	if baseLevel < 1 || baseLevel > 1000 {
		return fmt.Errorf("level precisa estar entre 1 e 1000. Exemplo: /train 5 351 391 -50")
	}
	if statLevel < 5 || statLevel > 1000 {
		return fmt.Errorf("skill precisa estar entre 5 e 1000. Exemplo: /train 5 351 391 -50")
	}
	if extraStats < -80 || extraStats > 126 {
		return fmt.Errorf("add precisa estar entre -80 e 126. Exemplo: /train 5 351 391 -50")
	}
	if attack < 4 || attack > 60 || attack == 6 || attack == 8 || attack == 10 || attack == 12 || attack == 14 {
		return fmt.Errorf("arma invalida. Use o ataque da arma de treino. Exemplo: /train 5 351 391 -50")
	}
	if targetEfficiency < 35 || targetEfficiency > 99 {
		return fmt.Errorf("eficiencia precisa estar entre 35 e 99. Exemplo: /train 5 351 391 -50 90")
	}

	return nil
}

func calculateRucoyTraining(baseLevel int, statLevel int, extraStats int, attack int, powertrain bool, targetEfficiency float64) RucoyTrainingResult {
	crit := 0.01
	critMulti := 1.05
	totalStat := statLevel + extraStats
	ticks := 10.0
	specMulti := 1.0
	mode := "AFK Train"
	if powertrain {
		ticks = 38
		specMulti = 1.5
		mode = "Powertrain"
	}

	minDamage := specMulti * (float64(baseLevel)/4 + float64(attack*totalStat)/20)
	maxDamage := specMulti * (float64(baseLevel)/4 + float64(attack*totalStat)/10)
	avgCritMulti := 1 + (critMulti-1)/2
	targetProb := 1 - math.Pow((100-targetEfficiency)/100, 1/ticks)

	result := RucoyTrainingResult{
		Mode:            mode,
		MinimumDuration: rucoyMinimumTrainingDurationSeconds,
	}

	for _, monster := range rucoyTrainingMonsters() {
		if powertrain && !monster.Powertrain {
			continue
		}

		prob := math.Min((1-crit)*(maxDamage-float64(monster.Defense))/(maxDamage-minDamage)+crit, 1)
		if targetProb < prob {
			finalProb := 100 - 100*math.Pow(1-prob, ticks)
			duration := rucoyTrainingDuration(monster, minDamage, maxDamage, crit, critMulti, avgCritMulti, prob)
			if duration <= 0 {
				continue
			}
			if duration < rucoyMinimumTrainingDurationSeconds {
				if result.BestShortDuration < duration {
					result.BestShortMonster = monster.Name
					result.BestShortEfficiency = finalProb
					result.BestShortDuration = duration
				}
				continue
			}

			if result.DurationSeconds < duration {
				result.Monster = monster.Name
				result.Efficiency = finalProb
				result.DurationSeconds = duration
				result.MaxDamage = int(math.Floor(maxDamage)) - monster.Defense
				result.MaxCriticalDamage = int(math.Floor(maxDamage*critMulti)) - monster.Defense
			} else if result.DurationSeconds == duration {
				result.Monster += " & " + monster.Name
			}
			continue
		}

		result.NextMonster = monster.Name
		result.RequiredStats = int(math.Ceil(
			(20*float64(monster.Defense)-20*float64(baseLevel)/4*specMulti)/
				(float64(attack)*specMulti*(2-(targetProb-crit)/(1-crit))),
		)) - totalStat
		result.StatsNeededFor1Damage = int(math.Ceil(
			10*((float64(1+monster.Defense)/specMulti)-float64(baseLevel)/4)/float64(attack),
		)) - totalStat
		break
	}

	return result
}

func rucoyTrainingDuration(monster RucoyTrainingMonster, minDamage float64, maxDamage float64, crit float64, critMulti float64, avgCritMulti float64, prob float64) float64 {
	monsterDefense := float64(monster.Defense)
	var damagePerSecond float64
	if minDamage < monsterDefense {
		damagePerSecond = crit*(maxDamage*avgCritMulti-monsterDefense) +
			(1-crit)*(maxDamage-monsterDefense)*prob/2
	} else {
		damagePerSecond = crit*(maxDamage*avgCritMulti-monsterDefense) +
			(1-crit)*(maxDamage+minDamage-2*monsterDefense)/2
	}

	if damagePerSecond <= 0 {
		return 0
	}

	return monster.HP / damagePerSecond
}

func writeRucoyTrainingResult(sb *strings.Builder, result RucoyTrainingResult, targetEfficiency float64) {
	fmt.Fprintf(sb, "%s:\n", result.Mode)
	if result.Monster == "" {
		fmt.Fprintf(sb, "Nenhum monstro viavel com %s+ e %.0f%%+ de eficiencia.\n", formatRucoyTrainingDuration(result.MinimumDuration), targetEfficiency)
		if result.BestShortMonster != "" {
			fmt.Fprintf(sb, "Melhor acima da eficiencia, mas ruim: %s\n", result.BestShortMonster)
			fmt.Fprintf(sb, "Ele morreria em media em %s.\n", formatRucoyTrainingDuration(result.BestShortDuration))
		}
		writeRucoyTrainingNextStep(sb, result, targetEfficiency)
		return
	}

	fmt.Fprintf(sb, "Melhor local: %s\n", result.Monster)
	fmt.Fprintf(sb, "Eficiencia estimada: %.1f%%\n", result.Efficiency)
	fmt.Fprintf(sb, "Tempo medio ate matar o mob: %s\n", formatRucoyTrainingDuration(result.DurationSeconds))
	if result.DurationSeconds > 450 {
		sb.WriteString("Obs: o mob pode exaurir antes, por volta de 07:30.\n")
	}
	fmt.Fprintf(sb, "Dano max: %d | crit max: %d\n", result.MaxDamage, result.MaxCriticalDamage)

	writeRucoyTrainingNextStep(sb, result, targetEfficiency)
}

func writeRucoyTrainingNextStep(sb *strings.Builder, result RucoyTrainingResult, targetEfficiency float64) {
	if result.NextMonster == "" {
		sb.WriteString("Proximo mob: nenhum acima na tabela atual.\n")
		return
	}

	fmt.Fprintf(sb, "Proximo mob: %s\n", result.NextMonster)
	if result.RequiredStats > 0 {
		fmt.Fprintf(sb, "Para avancar: falta +%d skill/add para %.0f%%+ de eficiencia.\n", result.RequiredStats, targetEfficiency)
		return
	}
	if result.StatsNeededFor1Damage > 0 {
		fmt.Fprintf(sb, "Para avancar: falta +%d skill/add para dar 1 dano max.\n", result.StatsNeededFor1Damage)
		return
	}

	sb.WriteString("Para avancar: voce ja esta perto; teste uma eficiencia alvo menor se quiser forcar esse mob.\n")
}

func writeRucoyTrainingWeaponAlternatives(sb *strings.Builder, currentAttack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64, powertrain bool) {
	alternatives := rucoyTrainingWeaponAlternatives(currentAttack, baseLevel, statLevel, extraStats, targetEfficiency, powertrain)
	mode := "AFK Train"
	if powertrain {
		mode = "Powertrain"
	}
	if len(alternatives) == 0 {
		fmt.Fprintf(sb, "Sugestoes: nem mudando so a arma de treino achei um %s 08:00+ nessa tabela.\n", mode)
		return
	}

	sb.WriteString("Sugestoes com arma de treino:\n")
	limit := min(len(alternatives), 5)
	for i := range limit {
		alternative := alternatives[i]
		fmt.Fprintf(
			sb,
			"- arma %d: %s por %s (%.1f%%)\n",
			alternative.Attack,
			alternative.Monster,
			formatRucoyTrainingDuration(alternative.DurationSeconds),
			alternative.Efficiency,
		)
	}
}

func rucoyTrainingWeaponAlternatives(currentAttack int, baseLevel int, statLevel int, extraStats int, targetEfficiency float64, powertrain bool) []RucoyTrainingAlternative {
	alternatives := make([]RucoyTrainingAlternative, 0)
	for _, attack := range rucoyTrainingWeaponAttacks() {
		if attack == currentAttack {
			continue
		}

		result := calculateRucoyTraining(baseLevel, statLevel, extraStats, attack, powertrain, targetEfficiency)
		if result.Monster == "" {
			continue
		}

		alternatives = append(alternatives, RucoyTrainingAlternative{
			Attack:          attack,
			Monster:         result.Monster,
			Efficiency:      result.Efficiency,
			DurationSeconds: result.DurationSeconds,
		})
	}

	sort.SliceStable(alternatives, func(i, j int) bool {
		return alternatives[i].DurationSeconds > alternatives[j].DurationSeconds
	})

	return alternatives
}

func formatRucoyTrainingDuration(seconds float64) string {
	totalSeconds := int(math.Round(seconds))
	minutes := totalSeconds / 60
	remainingSeconds := totalSeconds % 60

	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

func rucoyTrainingWeaponAttacks() []int {
	return []int{4, 5, 7, 9, 11, 13}
}

func rucoyTrainingMonsters() []RucoyTrainingMonster {
	return []RucoyTrainingMonster{
		{Name: "Rat Lv.1", Defense: 4, HP: 25, Powertrain: false},
		{Name: "Rat Lv.3", Defense: 7, HP: 35, Powertrain: false},
		{Name: "Crow Lv.6", Defense: 13, HP: 40, Powertrain: false},
		{Name: "Wolf Lv.9", Defense: 17, HP: 50, Powertrain: false},
		{Name: "Scorpion Lv.12", Defense: 18, HP: 50, Powertrain: false},
		{Name: "Cobra Lv.13", Defense: 18, HP: 50, Powertrain: false},
		{Name: "Worm Lv.14", Defense: 19, HP: 55, Powertrain: false},
		{Name: "Goblin Lv.15", Defense: 21, HP: 60, Powertrain: true},
		{Name: "Mummy Lv.25", Defense: 36, HP: 80, Powertrain: true},
		{Name: "Pharaoh Lv.35", Defense: 51, HP: 100, Powertrain: true},
		{Name: "Assassin Lv.45", Defense: 71, HP: 120, Powertrain: true},
		{Name: "Assassin Lv.50", Defense: 81, HP: 140, Powertrain: true},
		{Name: "Assassin Ninja Lv.55", Defense: 91, HP: 160, Powertrain: true},
		{Name: "Skeleton Archer Lv.80", Defense: 101, HP: 300, Powertrain: false},
		{Name: "Zombie Lv.65", Defense: 106, HP: 200, Powertrain: true},
		{Name: "Skeleton Lv.75", Defense: 121, HP: 300, Powertrain: true},
		{Name: "Skeleton Warrior Lv.90", Defense: 146, HP: 375, Powertrain: true},
		{Name: "Vampire Lv.100", Defense: 171, HP: 450, Powertrain: true},
		{Name: "Vampire Lv.110", Defense: 186, HP: 530, Powertrain: true},
		{Name: "Drow Ranger Lv.125", Defense: 191, HP: 600, Powertrain: false},
		{Name: "Drow Mage Lv.130", Defense: 191, HP: 600, Powertrain: false},
		{Name: "Drow Assassin Lv.120", Defense: 221, HP: 620, Powertrain: true},
		{Name: "Drow Sorceress Lv.140", Defense: 221, HP: 600, Powertrain: false},
		{Name: "Drow Fighter Lv.135", Defense: 246, HP: 680, Powertrain: true},
		{Name: "Lizard Archer Lv.160", Defense: 271, HP: 650, Powertrain: false},
		{Name: "Lizard Shaman Lv.170", Defense: 276, HP: 600, Powertrain: false},
		{Name: "Dead Eyes Lv.170", Defense: 276, HP: 600, Powertrain: false},
		{Name: "Lizard Warrior Lv.150", Defense: 301, HP: 680, Powertrain: true},
		{Name: "Djinn Lv.150", Defense: 301, HP: 640, Powertrain: true},
		{Name: "Lizard High Shaman Lv.190", Defense: 326, HP: 740, Powertrain: false},
		{Name: "Gargoyle Lv.190", Defense: 326, HP: 740, Powertrain: true},
		{Name: "Dragon Hatchling Lv.240", Defense: 331, HP: 10000, Powertrain: false},
		{Name: "Lizard Captain Lv.180", Defense: 361, HP: 815, Powertrain: true},
		{Name: "Dragon Lv.250", Defense: 501, HP: 20000, Powertrain: false},
		{Name: "Minotaur Lv.225", Defense: 511, HP: 4250, Powertrain: true},
		{Name: "Minotaur Lv.250", Defense: 601, HP: 5000, Powertrain: true},
		{Name: "Dragon Warden Lv.280", Defense: 626, HP: 30000, Powertrain: false},
		{Name: "Ice Elemental Lv.300", Defense: 676, HP: 40000, Powertrain: false},
		{Name: "Minotaur Lv.275", Defense: 691, HP: 5750, Powertrain: true},
		{Name: "Ice Dragon Lv.320", Defense: 726, HP: 45000, Powertrain: false},
		{Name: "Yeti Lv.350", Defense: 826, HP: 55000, Powertrain: false},
		{Name: "Lava Golem Lv.375", Defense: 900, HP: 65000, Powertrain: false},
		{Name: "Orthrus Lv.400", Defense: 1300, HP: 75000, Powertrain: false},
		{Name: "Demon Lv.450", Defense: 1550, HP: 100000, Powertrain: false},
	}
}
