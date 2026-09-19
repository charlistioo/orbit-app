package insight

import (
	"strings"
	"testing"
)

func TestGenerate_NoTrendClaimWithOneDayOfHistory(t *testing.T) {
	in := Inputs{
		Period:              "daily",
		PlanTotal:           50000,
		ActualTotal:         30000,
		DistinctHistoryDays: 1,
		TrendWindow: [3]DayCategoryTotals{
			nil, nil,
			{"Kopi": 30000},
		},
	}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "berturut-turut") {
			t.Fatalf("expected no trend claim with only 1 day of history, got: %q", msg)
		}
	}
}

func TestGenerate_TrendObservationFor3ConsecutiveRisingDays(t *testing.T) {
	in := Inputs{
		Period:              "daily",
		PlanTotal:           50000,
		ActualTotal:         30000,
		DistinctHistoryDays: 3,
		TrendWindow: [3]DayCategoryTotals{
			{"Kopi": 10000},
			{"Kopi": 20000},
			{"Kopi": 30000},
		},
	}

	insights := Generate(in)

	found := false
	for _, msg := range insights {
		if strings.Contains(msg, "Kopi") && strings.Contains(msg, "berturut-turut") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a trend observation for Kopi's 3 consecutive rising days, got: %v", insights)
	}
}

func TestGenerate_NoTrendWhenNotStrictlyRising(t *testing.T) {
	in := Inputs{
		Period:              "daily",
		DistinctHistoryDays: 3,
		ActualTotal:         1,
		TrendWindow: [3]DayCategoryTotals{
			{"Kopi": 30000},
			{"Kopi": 20000}, // dropped, not a rising trend
			{"Kopi": 30000},
		},
	}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "berturut-turut") {
			t.Fatalf("expected no trend claim when spend did not strictly rise each day, got: %q", msg)
		}
	}
}

func TestGenerate_MoodNeverStatedAsCause(t *testing.T) {
	in := Inputs{
		Period:             "daily",
		ActualTotal:        30000,
		PlanTotal:          20000,
		MoodBeforeSpending: true,
	}

	insights := Generate(in)

	found := false
	for _, msg := range insights {
		if strings.Contains(msg, "mood") {
			found = true
			causalPhrases := []string{"karena mood", "akibat mood", "disebabkan", "menyebabkan pengeluaran"}
			for _, phrase := range causalPhrases {
				if strings.Contains(msg, phrase) {
					t.Fatalf("mood insight must never state causation, got: %q", msg)
				}
			}
			if !strings.Contains(msg, "bukan berarti") && !strings.Contains(msg, "hanya kaitan waktu") {
				t.Fatalf("mood insight must explicitly frame itself as correlation, not causation, got: %q", msg)
			}
		}
	}
	if !found {
		t.Fatalf("expected a mood-related insight when MoodBeforeSpending is true, got: %v", insights)
	}
}

func TestGenerate_NoMoodInsightWhenNoTemporalCorrelation(t *testing.T) {
	in := Inputs{
		Period:             "daily",
		ActualTotal:        30000,
		PlanTotal:          20000,
		MoodBeforeSpending: false,
	}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "mood") {
			t.Fatalf("expected no mood insight when there was no temporal correlation, got: %q", msg)
		}
	}
}

func TestGenerate_NoDataMessageWhenNothingRecorded(t *testing.T) {
	in := Inputs{Period: "weekly"}

	insights := Generate(in)

	if len(insights) != 1 || !strings.Contains(insights[0], "Belum ada data") {
		t.Fatalf("expected a single no-data message, got: %v", insights)
	}
}

func TestGenerate_OverspendMessageWhenActualExceedsPlan(t *testing.T) {
	in := Inputs{Period: "daily", PlanTotal: 10000, ActualTotal: 15000}

	insights := Generate(in)

	if !strings.Contains(insights[0], "melebihi rencana") {
		t.Fatalf("expected overspend framing, got: %q", insights[0])
	}
}

func TestGenerate_BankCreditedAndApplied(t *testing.T) {
	in := Inputs{
		Period:       "daily",
		ActualTotal:  1,
		BankCredited: 5000,
		BankApplied:  2000,
	}

	insights := Generate(in)

	foundCredited, foundApplied := false, false
	for _, msg := range insights {
		if strings.Contains(msg, "masuk ke Consumption Bank") {
			foundCredited = true
		}
		if strings.Contains(msg, "digunakan untuk menutup") {
			foundApplied = true
		}
	}
	if !foundCredited {
		t.Fatalf("expected a credited insight, got: %v", insights)
	}
	if !foundApplied {
		t.Fatalf("expected an applied insight, got: %v", insights)
	}
}

func TestGenerate_NoBankInsightWhenNothingMoved(t *testing.T) {
	in := Inputs{Period: "daily", ActualTotal: 1}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "Consumption Bank") {
			t.Fatalf("expected no Consumption Bank insight with zero movement, got: %q", msg)
		}
	}
}

func TestGenerate_MostFrequentCategoryReported(t *testing.T) {
	in := Inputs{
		Period:            "weekly",
		ActualTotal:       1,
		CategoryFrequency: map[string]int{"Kopi": 5, "Makanan": 2},
	}

	insights := Generate(in)

	found := false
	for _, msg := range insights {
		if strings.Contains(msg, "paling sering dipakai") && strings.Contains(msg, "Kopi") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Kopi reported as most frequent category, got: %v", insights)
	}
}

func TestGenerate_NoMostFrequentCategoryOnTie(t *testing.T) {
	in := Inputs{
		Period:            "weekly",
		ActualTotal:       1,
		CategoryFrequency: map[string]int{"Kopi": 3, "Makanan": 3},
	}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "paling sering dipakai") {
			t.Fatalf("expected no most-frequent claim on a tie, got: %q", msg)
		}
	}
}

func TestGenerate_NoMostFrequentCategoryWithOnlyOneCategory(t *testing.T) {
	in := Inputs{
		Period:            "weekly",
		ActualTotal:       1,
		CategoryFrequency: map[string]int{"Kopi": 5},
	}

	insights := Generate(in)

	for _, msg := range insights {
		if strings.Contains(msg, "paling sering dipakai") {
			t.Fatalf("expected no most-frequent claim with only one category, got: %q", msg)
		}
	}
}
