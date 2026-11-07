package holding

import (
	"cmp"
	"sort"
	"strings"

	holdingmodel "keepnode/app/models/holding"
)

func documentFields(doc holdingmodel.SearchDocument) map[string]string {
	return map[string]string{
		"currency": doc.Currency, "exchange": doc.Exchange, "asset_class": doc.AssetClass,
		"sector": doc.Sector, "country": doc.Country,
	}
}

func searchDatabase(documents []holdingmodel.SearchDocument, p holdingmodel.SearchParams) *holdingmodel.SearchPage {
	filtered := []holdingmodel.SearchDocument{}
	for _, doc := range documents {
		fields := documentFields(doc)
		match := true
		for key, value := range map[string]string{
			"currency": p.Currency, "exchange": p.Exchange, "asset_class": p.AssetClass,
			"sector": p.Sector, "country": p.Country,
		} {
			if value != "" && fields[key] != value {
				match = false
				break
			}
		}
		text := strings.ToLower(strings.Join([]string{doc.Symbol, doc.Name, doc.Currency, doc.Exchange,
			doc.AssetClass, doc.Sector, doc.Country}, " "))
		for _, word := range strings.Fields(strings.ToLower(p.Query)) {
			if !strings.Contains(text, word) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, doc)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		a, b := filtered[i], filtered[j]
		order := 0
		switch p.Sort {
		case "quantity":
			order = cmp.Compare(a.Quantity, b.Quantity)
		case "buy_price":
			order = cmp.Compare(a.BuyPrice, b.BuyPrice)
		case "created_at":
			order = cmp.Compare(a.CreatedAt, b.CreatedAt)
		case "name":
			order = strings.Compare(a.Name, b.Name)
		default:
			order = strings.Compare(a.Symbol, b.Symbol)
		}
		if p.Direction == "desc" && p.Sort != "relevance" {
			order = -order
		}
		if order == 0 {
			order = strings.Compare(a.Symbol+a.ID, b.Symbol+b.ID)
		}
		return order < 0
	})
	page := newSearchPage(p, len(filtered), "database")
	sets := map[string]map[string]bool{}
	for _, key := range facetFields {
		sets[key] = map[string]bool{}
	}
	for _, doc := range filtered {
		for key, value := range documentFields(doc) {
			if value != "" {
				sets[key][value] = true
			}
		}
	}
	for key, set := range sets {
		page.Facets[key] = []string{}
		for value := range set {
			page.Facets[key] = append(page.Facets[key], value)
		}
		sort.Strings(page.Facets[key])
	}
	start := min((p.Page-1)*p.PerPage, len(filtered))
	page.Documents = filtered[start:min(start+p.PerPage, len(filtered))]
	return page
}
