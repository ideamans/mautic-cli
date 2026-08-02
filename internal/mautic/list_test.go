package mautic

import (
	"encoding/json"
	"testing"
)

// Mautic returns collections as an object keyed by record ID, and `total` as
// either a string or a number, depending on the endpoint. Both have to flatten
// to the same shape or every downstream consumer needs the special case.
func TestDecodePageObjectKeyedByID(t *testing.T) {
	raw := json.RawMessage(`{"total":"3","contacts":{"51":{"id":51},"7":{"id":7},"200":{"id":200}}}`)
	page, err := DecodePage(raw, "contacts")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
	// Numeric, not lexicographic: 7 < 51 < 200.
	var ids []int
	for _, item := range page.Items {
		var rec struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(item, &rec); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, rec.ID)
	}
	want := []int{7, 51, 200}
	for i := range want {
		if i >= len(ids) || ids[i] != want[i] {
			t.Fatalf("ids = %v, want %v", ids, want)
		}
	}
}

func TestDecodePageArray(t *testing.T) {
	raw := json.RawMessage(`{"total":2,"emails":[{"id":1},{"id":2}]}`)
	page, err := DecodePage(raw, "emails")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("got total=%d items=%d", page.Total, len(page.Items))
	}
}

// When the caller does not know the envelope key — which is the case for the
// raw `api` command and for endpoints this CLI does not wrap — fall back to the
// first collection alongside `total`.
func TestDecodePageFallsBackToTheFirstCollection(t *testing.T) {
	raw := json.RawMessage(`{"total":1,"somethingNew":[{"id":9}]}`)
	page, err := DecodePage(raw, "contacts")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if page.Key != "somethingNew" || len(page.Items) != 1 {
		t.Fatalf("got key=%q items=%d", page.Key, len(page.Items))
	}
}

// Without a `total`, the count of returned records is the only honest answer.
func TestDecodePageWithoutTotal(t *testing.T) {
	raw := json.RawMessage(`{"stats":[{"id":1},{"id":2},{"id":3}]}`)
	page, err := DecodePage(raw, "stats")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
}

func TestParseCondition(t *testing.T) {
	c, err := ParseCondition("email:like:%@example.com")
	if err != nil {
		t.Fatalf("ParseCondition: %v", err)
	}
	if c.Col != "email" || c.Expr != "like" || c.Val != "%@example.com" {
		t.Fatalf("got %+v", c)
	}

	// A URL is a value with colons in it, and splitting on every colon would
	// mangle it.
	c, err = ParseCondition("url:eq:https://example.com/a")
	if err != nil {
		t.Fatalf("ParseCondition: %v", err)
	}
	if c.Val != "https://example.com/a" {
		t.Errorf("Val = %q, want the whole URL", c.Val)
	}

	// isNull takes no value.
	if _, err := ParseCondition("email:isNull"); err != nil {
		t.Errorf("isNull should not require a value: %v", err)
	}
	// Everything else does.
	if _, err := ParseCondition("email:eq"); err == nil {
		t.Error("eq without a value should be rejected")
	}
	if _, err := ParseCondition("nonsense"); err == nil {
		t.Error("a clause with no expression should be rejected")
	}
}

func TestParseOrder(t *testing.T) {
	o, err := ParseOrder("date_hit:desc")
	if err != nil {
		t.Fatalf("ParseOrder: %v", err)
	}
	if o.Col != "date_hit" || o.Dir != "DESC" {
		t.Fatalf("got %+v", o)
	}
	if o, err := ParseOrder("id"); err != nil || o.Dir != "ASC" {
		t.Errorf("a bare column should default to ASC: %+v %v", o, err)
	}
	if _, err := ParseOrder("id:sideways"); err == nil {
		t.Error("an invalid direction should be rejected")
	}
}

func TestListParamsEncodesAdvancedFilters(t *testing.T) {
	p := ListParams{
		Limit:   10,
		Minimal: true,
		Where:   []Condition{{Col: "phone", Expr: "in", Val: "111,222"}},
		Order:   []OrderClause{{Col: "date_hit", Dir: "DESC"}},
	}
	v := p.Values()
	cases := map[string]string{
		"limit":          "10",
		"minimal":        "true",
		"where[0][col]":  "phone",
		"where[0][expr]": "in",
		"where[0][val]":  "111,222",
		"order[0][col]":  "date_hit",
		"order[0][dir]":  "DESC",
	}
	for key, want := range cases {
		if got := v.Get(key); got != want {
			t.Errorf("%s = %q, want %q", key, got, want)
		}
	}
	// Zero values must not be sent: start=0 is Mautic's default and sending an
	// empty search would filter everything out on some endpoints.
	if _, ok := v["start"]; ok {
		t.Error("start=0 should be omitted")
	}
	if _, ok := v["search"]; ok {
		t.Error("an empty search should be omitted")
	}
}

// /contacts/list/fields returns the collection as the entire response body,
// keyed by record ID, with no wrapper key and no total. Without a case for
// this, the first record is mistaken for the collection and its properties come
// out as rows.
func TestDecodePageBareCollection(t *testing.T) {
	raw := json.RawMessage(`{
	  "1":  {"id":1,"alias":"title","label":"Title"},
	  "10": {"id":10,"alias":"fax","label":"Fax"},
	  "2":  {"id":2,"alias":"firstname","label":"First Name"}
	}`)
	page, err := DecodePage(raw, "fields")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if len(page.Items) != 3 {
		t.Fatalf("got %d items, want 3", len(page.Items))
	}
	if page.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Total)
	}
	var first struct {
		Alias string `json:"alias"`
	}
	if err := json.Unmarshal(page.Items[0], &first); err != nil {
		t.Fatal(err)
	}
	// Numeric key order: 1, 2, 10.
	if first.Alias != "title" {
		t.Errorf("first item alias = %q, want title", first.Alias)
	}
}

// A normal envelope must not be mistaken for a bare collection.
func TestDecodePageDoesNotTreatEnvelopesAsBare(t *testing.T) {
	raw := json.RawMessage(`{"total":1,"contacts":{"5":{"id":5}}}`)
	page, err := DecodePage(raw, "contacts")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if len(page.Items) != 1 || page.Key != "contacts" {
		t.Fatalf("got key=%q items=%d", page.Key, len(page.Items))
	}
}

// Nor must a mixed object such as /data's {"success":1,"types":{…}}.
func TestDecodePageIgnoresMixedObjects(t *testing.T) {
	raw := json.RawMessage(`{"success":1,"types":{"a":{"x":1},"b":{"x":2}}}`)
	page, err := DecodePage(raw, "types")
	if err != nil {
		t.Fatalf("DecodePage: %v", err)
	}
	if page.Key != "types" || len(page.Items) != 2 {
		t.Fatalf("got key=%q items=%d", page.Key, len(page.Items))
	}
}
