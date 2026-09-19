// Package insight contains the pure, rule-based text-generation logic
// for ORBIT's insight engine. It has no database dependency - callers
// (backend/internal/store) fetch the aggregates it needs and hand them
// in as Inputs, which keeps the actual rules (what counts as "enough
// history", what counts as a trend, how mood is phrased) fully unit
// testable without a database.
package insight

import "fmt"

// DayCategoryTotals maps a category name to its total spend on one day -
// used only for the 3-day trend window.
type DayCategoryTotals map[string]float64

// Inputs is everything Generate needs to produce insight text for one
// period (a single day, a week, or a month).
type Inputs struct {
	Period      string // "daily" | "weekly" | "monthly"
	PlanTotal   float64
	ActualTotal float64

	// DistinctHistoryDays is how many distinct calendar days (anywhere
	// in the user's history up to and including the period) have at
	// least one recorded transaction. A trend claim requires at least 3 -
	// a user with only 1 day of data must never receive one.
	DistinctHistoryDays int

	// TrendWindow holds the 3 most recent distinct days' per-category
	// totals, oldest first (index 0 = 2 days before reference, index 2 =
	// reference day). A category strictly increasing across all three
	// entries (and non-zero on the first) is reported as trending. Only
	// consulted when DistinctHistoryDays >= 3.
	TrendWindow [3]DayCategoryTotals

	// MoodBeforeSpending is true when a mood entry was logged earlier in
	// the period than at least one transaction. It is phrased strictly as
	// a temporal observation ("mood was logged, then spending followed"),
	// never as a cause of the spending.
	MoodBeforeSpending bool

	// BankCredited/BankApplied are the Consumption Bank ledger movements
	// recorded within the period - credited (underspend recognized) and
	// applied (balance used to cover an overspend), each already netted
	// to a single non-negative amount for the period.
	BankCredited float64
	BankApplied  float64

	// CategoryFrequency maps category name -> number of transactions in
	// the period. Used only to name the single most-frequently-used
	// category - and only when there's more than one category to compare
	// and no tie for first place, so a lone category (or an ambiguous
	// tie) never gets reported as "most frequent", per the same
	// never-invent-a-conclusion rule as the trend check.
	CategoryFrequency map[string]int
}

// Generate produces the insight strings for one period, following three
// rules: (1) never claim a multi-day trend without at least 3 distinct
// days of history, (2) surface a trend observation whenever spend in one
// category has strictly risen for 3 consecutive days, (3) never phrase
// mood as the cause of spending - only as something that happened first.
func Generate(in Inputs) []string {
	insights := []string{}

	if in.PlanTotal == 0 && in.ActualTotal == 0 {
		insights = append(insights, periodNoDataMessage(in.Period))
	} else {
		insights = append(insights, planVsActualMessage(in.Period, in.PlanTotal, in.ActualTotal))
	}

	if in.DistinctHistoryDays >= 3 {
		for _, category := range risingCategories(in.TrendWindow) {
			insights = append(insights, fmt.Sprintf(
				"Pengeluaran kategori %s naik 3 hari berturut-turut - ini pola yang layak diperhatikan.",
				category,
			))
		}
	}

	if in.MoodBeforeSpending {
		insights = append(insights,
			"Ada perubahan mood yang tercatat sebelum beberapa transaksi - ini hanya kaitan waktu, bukan berarti mood tersebut menjadi penyebab pengeluaran.",
		)
	}

	if in.BankCredited > 0 {
		insights = append(insights, fmt.Sprintf(
			"%s masuk ke Consumption Bank dari sisa anggaran yang tidak terpakai.",
			formatRupiah(in.BankCredited),
		))
	}
	if in.BankApplied > 0 {
		insights = append(insights, fmt.Sprintf(
			"%s dari Consumption Bank digunakan untuk menutup pengeluaran yang melebihi rencana.",
			formatRupiah(in.BankApplied),
		))
	}

	if topCategory, ok := mostFrequentCategory(in.CategoryFrequency); ok {
		insights = append(insights, fmt.Sprintf(
			"Kategori yang paling sering dipakai: %s.",
			topCategory,
		))
	}

	return insights
}

// mostFrequentCategory returns the category with the highest transaction
// count, requiring at least 2 categories to compare and no tie for first
// place - a single category, or an exact tie, isn't reported, since
// neither actually supports a "most frequent" conclusion.
func mostFrequentCategory(frequency map[string]int) (string, bool) {
	if len(frequency) < 2 {
		return "", false
	}

	topName := ""
	topCount := 0
	tied := false
	for name, count := range frequency {
		switch {
		case count > topCount:
			topName, topCount, tied = name, count, false
		case count == topCount:
			tied = true
		}
	}
	if tied {
		return "", false
	}
	return topName, true
}

func periodNoDataMessage(period string) string {
	switch period {
	case "weekly":
		return "Belum ada data pengeluaran untuk minggu ini."
	case "monthly":
		return "Belum ada data pengeluaran untuk bulan ini."
	default:
		return "Belum ada data pengeluaran untuk hari ini."
	}
}

func planVsActualMessage(period string, planTotal, actualTotal float64) string {
	scope := "hari ini"
	if period == "weekly" {
		scope = "minggu ini"
	} else if period == "monthly" {
		scope = "bulan ini"
	}

	if planTotal == 0 {
		return fmt.Sprintf("Belum ada rencana yang ditetapkan untuk %s, tercatat pengeluaran %s.", scope, formatRupiah(actualTotal))
	}
	if actualTotal > planTotal {
		return fmt.Sprintf("Pengeluaran %s (%s) melebihi rencana (%s).", scope, formatRupiah(actualTotal), formatRupiah(planTotal))
	}
	return fmt.Sprintf("Pengeluaran %s (%s) masih dalam rencana (%s).", scope, formatRupiah(actualTotal), formatRupiah(planTotal))
}

func formatRupiah(amount float64) string {
	return fmt.Sprintf("Rp%.0f", amount)
}

// risingCategories returns the names of categories whose totals strictly
// increased across all three entries of the trend window (and were
// non-zero on the first day, so a category that merely appeared for the
// first time isn't reported as "rising").
func risingCategories(window [3]DayCategoryTotals) []string {
	if window[0] == nil || window[1] == nil || window[2] == nil {
		return nil
	}

	trending := []string{}
	for category, day1 := range window[0] {
		if day1 <= 0 {
			continue
		}
		day2, ok2 := window[1][category]
		day3, ok3 := window[2][category]
		if ok2 && ok3 && day2 > day1 && day3 > day2 {
			trending = append(trending, category)
		}
	}
	return trending
}
