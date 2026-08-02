package mautic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Condition is one advanced `where` clause. Mautic passes these through to
// Doctrine, so expr is a Doctrine expression name: eq, neq, gt, gte, lt, lte,
// like, notLike, in, notIn, isNull, isNotNull, between.
type Condition struct {
	Col  string
	Expr string
	Val  string
}

// OrderClause is one advanced `order` statement.
type OrderClause struct {
	Col string
	Dir string
}

// ListParams are the query parameters every Mautic collection endpoint accepts.
type ListParams struct {
	Search string
	Start  int
	Limit  int
	// OrderBy is a *column* name, not the camelCase property name from the
	// response: dateIdentified is date_identified, modifiedByUser is
	// modified_by_user. Getting this wrong is a silent no-op, not an error.
	OrderBy       string
	OrderByDir    string
	PublishedOnly bool
	Minimal       bool
	Where         []Condition
	Order         []OrderClause
	// Extra carries endpoint-specific parameters such as dateFrom/dateTo.
	Extra url.Values
}

// Values renders the parameters as a query string.
func (p ListParams) Values() url.Values {
	v := url.Values{}
	for key, vals := range p.Extra {
		for _, val := range vals {
			v.Add(key, val)
		}
	}
	if p.Search != "" {
		v.Set("search", p.Search)
	}
	if p.Start > 0 {
		v.Set("start", strconv.Itoa(p.Start))
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.OrderBy != "" {
		v.Set("orderBy", p.OrderBy)
	}
	if p.OrderByDir != "" {
		v.Set("orderByDir", p.OrderByDir)
	}
	if p.PublishedOnly {
		v.Set("publishedOnly", "true")
	}
	if p.Minimal {
		v.Set("minimal", "true")
	}
	for i, c := range p.Where {
		v.Set(fmt.Sprintf("where[%d][col]", i), c.Col)
		v.Set(fmt.Sprintf("where[%d][expr]", i), c.Expr)
		if c.Val != "" {
			v.Set(fmt.Sprintf("where[%d][val]", i), c.Val)
		}
	}
	for i, o := range p.Order {
		v.Set(fmt.Sprintf("order[%d][col]", i), o.Col)
		v.Set(fmt.Sprintf("order[%d][dir]", i), o.Dir)
	}
	return v
}

// ParseCondition reads a where clause written as "col:expr:val", which is what
// the --where flag takes. The value may itself contain colons.
func ParseCondition(s string) (Condition, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) < 2 {
		return Condition{}, fmt.Errorf("could not parse --where %q; use col:expr[:value], for example email:like:%%@example.com", s)
	}
	c := Condition{Col: parts[0], Expr: parts[1]}
	if len(parts) == 3 {
		c.Val = parts[2]
	}
	switch c.Expr {
	case "isNull", "isNotNull":
		// No value expected.
	default:
		if c.Val == "" {
			return Condition{}, fmt.Errorf("--where %q needs a value: col:expr:value", s)
		}
	}
	return c, nil
}

// ParseOrder reads an order clause written as "col:dir".
func ParseOrder(s string) (OrderClause, error) {
	parts := strings.SplitN(s, ":", 2)
	o := OrderClause{Col: parts[0], Dir: "ASC"}
	if len(parts) == 2 {
		o.Dir = strings.ToUpper(parts[1])
	}
	if o.Dir != "ASC" && o.Dir != "DESC" {
		return OrderClause{}, fmt.Errorf("--order %q: direction must be asc or desc", s)
	}
	return o, nil
}

// Page is one decoded collection response.
//
// Mautic is inconsistent here: most endpoints return the collection as an
// object keyed by record ID, a few return a plain array, and `total` arrives as
// either a string or a number. All of that is normalised away.
type Page struct {
	Total int
	Key   string
	Items []json.RawMessage
}

// DecodePage pulls the collection out of a list response. preferKey names the
// expected collection key; when it is absent the first array-or-object value
// alongside `total` is used, which keeps the raw `api` command working against
// endpoints this CLI does not wrap.
func DecodePage(raw json.RawMessage, preferKey string) (*Page, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("the response was not a JSON object: %w", err)
	}

	page := &Page{Key: preferKey}
	if t, ok := envelope["total"]; ok {
		page.Total = decodeLooseInt(t)
	}

	collection, ok := envelope[preferKey]
	if !ok && isBareCollection(envelope) {
		// A few endpoints — /contacts/list/fields is the one that bites —
		// return the collection as the entire response body, keyed by record
		// ID, with no wrapper and no total. Without this case the first record
		// gets mistaken for the collection and its properties come out as rows.
		collection, page.Key = raw, ""
		ok = true
	}
	if !ok {
		for _, key := range sortedMapKeys(envelope) {
			if key == "total" {
				continue
			}
			if looksLikeCollection(envelope[key]) {
				collection, page.Key = envelope[key], key
				ok = true
				break
			}
		}
	}
	if !ok {
		return page, nil
	}

	items, err := decodeCollection(collection)
	if err != nil {
		return nil, err
	}
	page.Items = items
	if _, hasTotal := envelope["total"]; !hasTotal {
		page.Total = len(items)
	}
	return page, nil
}

// decodeCollection flattens either an array or an ID-keyed object into a stable
// slice. ID-keyed objects are sorted numerically so output does not reshuffle
// between runs.
func decodeCollection(raw json.RawMessage) ([]json.RawMessage, error) {
	trimmed := strings.TrimSpace(string(raw))
	if strings.HasPrefix(trimmed, "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, err
		}
		return arr, nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	keys := sortedMapKeys(obj)
	sort.SliceStable(keys, func(i, j int) bool {
		ni, erri := strconv.Atoi(keys[i])
		nj, errj := strconv.Atoi(keys[j])
		if erri == nil && errj == nil {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	items := make([]json.RawMessage, 0, len(keys))
	for _, k := range keys {
		items = append(items, obj[k])
	}
	return items, nil
}

// isBareCollection reports whether the response body is itself the collection:
// no `total`, at least one entry, and every value an object. The `total` check
// is what keeps a normal envelope such as {"total":1,"contacts":{…}} out, and
// requiring every value to be an object keeps {"success":1,"types":{…}} out.
func isBareCollection(envelope map[string]json.RawMessage) bool {
	if len(envelope) == 0 {
		return false
	}
	if _, hasTotal := envelope["total"]; hasTotal {
		return false
	}
	for _, v := range envelope {
		if !strings.HasPrefix(strings.TrimSpace(string(v)), "{") {
			return false
		}
	}
	return true
}

func looksLikeCollection(raw json.RawMessage) bool {
	t := strings.TrimSpace(string(raw))
	return strings.HasPrefix(t, "[") || strings.HasPrefix(t, "{")
}

// decodeLooseInt accepts 12 and "12" alike; Mautic sends `total` both ways
// depending on the endpoint.
func decodeLooseInt(raw json.RawMessage) int {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
			return n
		}
	}
	return 0
}

func sortedMapKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ListAll walks a collection endpoint until every record has been fetched.
//
// pageSize bounds one request; Mautic's own default is 30 and instances cap it,
// so a large --limit is not a substitute for paging.
func (c *Client) ListAll(ctx context.Context, path string, params ListParams, key string, pageSize, max int) (*Page, error) {
	if pageSize <= 0 {
		pageSize = 100
	}
	out := &Page{Key: key}
	start := params.Start
	for {
		p := params
		p.Start = start
		p.Limit = pageSize
		if max > 0 {
			if remaining := max - len(out.Items); remaining < pageSize {
				p.Limit = remaining
			}
		}
		raw, err := c.Get(ctx, path, p.Values())
		if err != nil {
			return nil, err
		}
		page, err := DecodePage(raw, key)
		if err != nil {
			return nil, err
		}
		if out.Key == "" {
			out.Key = page.Key
		}
		out.Total = page.Total
		out.Items = append(out.Items, page.Items...)

		start += len(page.Items)
		switch {
		case len(page.Items) == 0:
			return out, nil
		case max > 0 && len(out.Items) >= max:
			out.Items = out.Items[:max]
			return out, nil
		case page.Total > 0 && start >= page.Total:
			return out, nil
		}
	}
}
