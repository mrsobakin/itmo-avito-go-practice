package domain

type Plan string

const (
	PlanBasic Plan = "basic"
	PlanPro   Plan = "pro"
	PlanMax   Plan = "max"
)

var planRank = map[Plan]int{
	PlanBasic: 1,
	PlanPro:   2,
	PlanMax:   3,
}

func ParsePlan(s string) (Plan, error) {
	p := Plan(s)
	if _, ok := planRank[p]; !ok {
		return "", ErrUnknownPlan
	}
	return p, nil
}

func (p Plan) IsUpgradeFrom(other Plan) bool   { return planRank[p] > planRank[other] }
func (p Plan) IsDowngradeFrom(other Plan) bool { return planRank[p] < planRank[other] }
