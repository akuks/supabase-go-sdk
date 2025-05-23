package supabasego

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Table provides CRUD operations for a specific Supabase table.
type Table struct {
	client     *Client
	tableName  string
	filters    []Filter
	orders     []order
	limit      int
	offset     int
	selectCols []string
}

// Filter interface and types
type Filter interface {
	toQuery() string
}

type simpleFilter struct {
	field string
	op    string
	value interface{}
}

func (f simpleFilter) toQuery() string {
	if f.value == nil {
		return fmt.Sprintf("%s.is.null", f.field)
	}
	if f.op == "in" {
		return fmt.Sprintf("%s.in.%v", f.field, f.value)
	}
	return fmt.Sprintf("%s.%s.%v", f.field, f.op, f.value)
}

type groupFilter struct {
	operator string // "and" or "or"
	filters  []Filter
}

func (g groupFilter) toQuery() string {
	var parts []string
	for _, f := range g.filters {
		parts = append(parts, f.toQuery())
	}
	return fmt.Sprintf("%s(%s)", g.operator, strings.Join(parts, ","))
}

// Filter constructors
func Eq(field string, value interface{}) Filter {
	return simpleFilter{field, "eq", value}
}
func NotEq(field string, value interface{}) Filter {
	return simpleFilter{field, "neq", value}
}
func Gt(field string, value interface{}) Filter {
	return simpleFilter{field, "gt", value}
}
func Lt(field string, value interface{}) Filter {
	return simpleFilter{field, "lt", value}
}
func Gte(field string, value interface{}) Filter {
	return simpleFilter{field, "gte", value}
}
func Lte(field string, value interface{}) Filter {
	return simpleFilter{field, "lte", value}
}
func Like(field string, pattern string) Filter {
	return simpleFilter{field, "like", pattern}
}
func ILike(field string, pattern string) Filter {
	return simpleFilter{field, "ilike", pattern}
}
func In(field string, values []interface{}) Filter {
	var strVals []string
	for _, v := range values {
		strVals = append(strVals, fmt.Sprintf("%v", v))
	}
	joined := strings.Join(strVals, ",")
	return simpleFilter{field, "in", fmt.Sprintf("(%s)", joined)}
}
func And(filters ...Filter) Filter {
	return groupFilter{"and", filters}
}
func Or(filters ...Filter) Filter {
	return groupFilter{"or", filters}
}

// filter, order, and other query option types will be defined here.
type order struct {
	field     string
	direction string // "asc" or "desc"
}

// Table returns a Table instance for the given table name.
func (c *Client) Table(name string) *Table {
	return &Table{
		client:    c,
		tableName: name,
	}
}

// AddFilter allows adding a filter (for AND/OR/nested support)
func (t *Table) AddFilter(f Filter) *Table {
	t.filters = append(t.filters, f)
	return t
}

// Keep Eq, Gt, etc. for backward compatibility
func (t *Table) Eq(field string, value interface{}) *Table { return t.AddFilter(Eq(field, value)) }
func (t *Table) NotEq(field string, value interface{}) *Table { return t.AddFilter(NotEq(field, value)) }
func (t *Table) Gt(field string, value interface{}) *Table { return t.AddFilter(Gt(field, value)) }
func (t *Table) Lt(field string, value interface{}) *Table { return t.AddFilter(Lt(field, value)) }
func (t *Table) Gte(field string, value interface{}) *Table { return t.AddFilter(Gte(field, value)) }
func (t *Table) Lte(field string, value interface{}) *Table { return t.AddFilter(Lte(field, value)) }
func (t *Table) Like(field string, pattern string) *Table { return t.AddFilter(Like(field, pattern)) }
func (t *Table) ILike(field string, pattern string) *Table { return t.AddFilter(ILike(field, pattern)) }
func (t *Table) In(field string, values []interface{}) *Table { return t.AddFilter(In(field, values)) }

// And/Or as chainable methods
func (t *Table) And(filters ...Filter) *Table { return t.AddFilter(And(filters...)) }
func (t *Table) Or(filters ...Filter) *Table { return t.AddFilter(Or(filters...)) }

// Limit sets the maximum number of records to return.
func (t *Table) Limit(n int) *Table {
	t.limit = n
	return t
}

// OrderBy adds an order clause to the query (direction should be "asc" or "desc").
func (t *Table) OrderBy(field, direction string) *Table {
	dir := strings.ToLower(direction)
	if dir != "asc" && dir != "desc" {
		dir = "asc"
	}
	t.orders = append(t.orders, order{field: field, direction: dir})
	return t
}

// Offset sets the number of records to skip.
func (t *Table) Offset(n int) *Table {
	t.offset = n
	return t
}

// SelectColumns sets the columns to fetch.
func (t *Table) SelectColumns(cols ...string) *Table {
	t.selectCols = cols
	return t
}

// Select fetches records from the table into dest (must be a pointer to a slice).
func (t *Table) Select(dest interface{}, jwtToken string) error {
	// Build the query string
	params := url.Values{}
	for _, f := range t.filters {
		params.Add("or", f.toQuery())
	}
	if t.limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", t.limit))
	}
	if t.offset > 0 {
		params.Add("offset", fmt.Sprintf("%d", t.offset))
	}
	if len(t.orders) > 0 {
		var orderParams []string
		for _, o := range t.orders {
			orderParams = append(orderParams, fmt.Sprintf("%s.%s", o.field, o.direction))
		}
		params.Add("order", strings.Join(orderParams, ","))
	}
	if len(t.selectCols) > 0 {
		params.Add("select", strings.Join(t.selectCols, ","))
	} else {
		params.Add("select", "*")
	}

	endpoint := fmt.Sprintf("%s%s/%s", t.client.BaseURL, REST_URL, t.tableName)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", t.client.APIKey)
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase: select failed: %s", string(body))
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

// Insert inserts one or more records into the table.
func (t *Table) Insert(records interface{}, jwtToken string) error {
	endpoint := fmt.Sprintf("%s%s/%s", t.client.BaseURL, REST_URL, t.tableName)
	b, err := json.Marshal(records)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", t.client.APIKey)
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation") // Return inserted rows

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase: insert failed: %s", string(body))
	}
	return nil
}

// Update updates records matching filters with given values.
func (t *Table) Update(values map[string]interface{}, jwtToken string) error {
	params := url.Values{}
	for _, f := range t.filters {
		params.Add("or", f.toQuery())
	}
	endpoint := fmt.Sprintf("%s%s/%s", t.client.BaseURL, REST_URL, t.tableName)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	b, err := json.Marshal(values)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PATCH", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", t.client.APIKey)
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation") // Return updated rows

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase: update failed: %s", string(body))
	}
	return nil
}

// Delete deletes records matching filters from the table.
func (t *Table) Delete(jwtToken string) error {
	params := url.Values{}
	for _, f := range t.filters {
		params.Add("or", f.toQuery())
	}
	endpoint := fmt.Sprintf("%s%s/%s", t.client.BaseURL, REST_URL, t.tableName)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", t.client.APIKey)
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	req.Header.Set("Prefer", "return=representation") // Return deleted rows

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase: delete failed: %s", string(body))
	}
	return nil
}
