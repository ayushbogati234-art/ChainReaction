package risk

import (
	"math"
	"strings"
)

// cvss3 accepts complete v3.0/v3.1 vectors and calculates BASE severity only.
// Metric values, equations and integer rounding: FIRST v3.1 sections 7 and A.
// https://www.first.org/cvss/v3.1/specification-document
func cvss3(vector string) (float64, string, bool) {
	parts := strings.Split(vector, "/")
	if len(parts) < 9 || (parts[0] != "CVSS:3.0" && parts[0] != "CVSS:3.1") {
		return 0, "", false
	}
	allowed := map[string]string{"AV": "NALP", "AC": "LH", "PR": "NLH", "UI": "NR", "S": "UC", "C": "NLH", "I": "NLH", "A": "NLH", "E": "XUPFH", "RL": "XOTWU", "RC": "XURC", "CR": "XLMH", "IR": "XLMH", "AR": "XLMH", "MAV": "XNALP", "MAC": "XLH", "MPR": "XNLH", "MUI": "XNR", "MS": "XUC", "MC": "XNLH", "MI": "XNLH", "MA": "XNLH"}
	m := map[string]string{}
	for _, part := range parts[1:] {
		key, value, ok := strings.Cut(part, ":")
		if !ok || len(value) != 1 || !strings.Contains(allowed[key], value) || m[key] != "" {
			return 0, "", false
		}
		m[key] = value
	}
	for _, key := range []string{"AV", "AC", "PR", "UI", "S", "C", "I", "A"} {
		if m[key] == "" {
			return 0, "", false
		}
	}
	av := map[string]float64{"N": .85, "A": .62, "L": .55, "P": .2}[m["AV"]]
	ac := map[string]float64{"L": .77, "H": .44}[m["AC"]]
	pr := map[string]float64{"N": .85, "L": .62, "H": .27}[m["PR"]]
	if m["S"] == "C" {
		pr = map[string]float64{"N": .85, "L": .68, "H": .5}[m["PR"]]
	}
	ui := map[string]float64{"N": .85, "R": .62}[m["UI"]]
	impactValue := map[string]float64{"N": 0, "L": .22, "H": .56}
	iss := 1 - (1-impactValue[m["C"]])*(1-impactValue[m["I"]])*(1-impactValue[m["A"]])
	impact := 6.42 * iss
	if m["S"] == "C" {
		impact = 7.52*(iss-.029) - 3.25*math.Pow(iss-.02, 15)
	}
	if impact <= 0 {
		return 0, m["E"], true
	}
	base := impact + 8.22*av*ac*pr*ui
	if m["S"] == "C" {
		base *= 1.08
	}
	// FIRST Appendix A: remove floating point noise before rounding upwards.
	integer := int(math.Round(math.Min(base, 10) * 100000))
	return float64((integer+9999)/10000) / 10, m["E"], true
}
