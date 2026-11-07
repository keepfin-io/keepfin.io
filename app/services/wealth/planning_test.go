package wealth

import (
	"github.com/google/uuid"
	portfolioservice "keepnode/app/services/portfolio"
	"testing"
)

func TestFireProjectionUsesStoredAssumptionsAndUnknowns(t *testing.T) {
	p := DefaultPreferences(uuid.New())
	p.AnnualExpenses = 960
	p.WithdrawalRate = 4
	p.AnnualReturn = 0
	p.MonthlyContribution = 100
	p.HorizonYears = 20
	net := 12000.
	r := &portfolioservice.PortfolioReport{Currency: "USD", Complete: true}
	r.Summary.NetWorth = &net
	result := FireProjection(r, p)
	if result.Target == nil || *result.Target != 24000 || result.YearsToTarget == nil || *result.YearsToTarget != 10 || len(result.Projection) != 21 {
		t.Fatalf("projection %+v", result)
	}
	if result.Projection[1].Value != 13200 {
		t.Fatal("monthly contribution arithmetic")
	}
	r.Complete = false
	result = FireProjection(r, p)
	if result.Investable != nil || result.Target != nil || len(result.Projection) != 0 {
		t.Fatal("incomplete data projected")
	}
}
func TestRiskChecksRespectDisabledRulesAndCashCoverage(t *testing.T) {
	p := DefaultPreferences(uuid.New())
	p.EmergencyFund = 1000
	p.MonthlyExpenses = 100
	p.DisabledRules = []string{"asset"}
	r := &portfolioservice.PortfolioReport{Currency: "USD", Complete: true, Allocations: map[string][]portfolioservice.Allocation{}}
	r.Summary.Cash = 200
	result := RiskChecks(r, p)
	for _, rule := range result.Rules {
		if rule.ID == "emergency" && (rule.Value == nil || *rule.Value != 2 || rule.Status != "warning") {
			t.Fatalf("emergency %+v", rule)
		}
		if rule.ID == "asset" && rule.Enabled {
			t.Fatal("disabled rule enabled")
		}
	}
	r.Complete = false
	for _, rule := range RiskChecks(r, p).Rules {
		if rule.Value != nil || rule.Status != "unknown" {
			t.Fatal("missing data not preserved")
		}
	}
}
