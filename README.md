# supabasego

A Go SDK for interacting with the Supabase REST API, inspired by the official JS/Python clients.

## Features
- Core API client
- Generic table CRUD operations (insert, select, update, delete)
- Chainable query builder for filters, ordering, pagination
- JWT and API key support for RLS
- **Automatic decoding of inserted rows with DB-generated fields (UUIDs, timestamps, etc.)**
- **Correct handling of nullable fields and nil values in filters and inserts**

## Getting Started

```go
import "yourmodule/supabasego"

cfg := supabasego.Config{
    BaseURL: "https://<project>.supabase.co",
    APIKey:  "<service_or_anon_key>",
}
client := supabasego.NewClient(cfg)
```

## Generic Table CRUD

### Usage Examples

### Insert (with DB-generated fields)
```go
// Insert a new agent/tenant; let DB generate UUIDs and timestamps
newAgent := Agent{
    Name: "Example Agent",
    // ID: nil (omit), DB will generate
}
var inserted []Agent
err := client.Table("agents").Insert(&[]Agent{newAgent}, jwtToken)
if err != nil {
    // handle error
}
// inserted[0] now contains all DB-generated fields (id, created_at, etc.)
```

### Select
```go
// Fetch up to 10 tenants for a given user (with RLS)
var tenants []Tenant
err := client.Table("tenants").
    Eq("user_id", userID).
    Limit(10).
    Select(&tenants, jwtToken)
if err != nil {
    // handle error
}
for _, t := range tenants {
    fmt.Println(t.Name)
}
```

### Update
```go
// Update the name of a tenant by ID
err := client.Table("tenants").
    Eq("id", "t1").
    Update(map[string]interface{}{ "name": "New Name" }, jwtToken)
if err != nil {
    // handle error
}
```

### Delete
```go
// Delete a tenant by ID
err := client.Table("tenants").
    Eq("id", "t1").
    Delete(jwtToken)
if err != nil {
    // handle error
}
```

- Use `.Eq()` to filter, `.Limit()` to restrict results.
- Pass a JWT token for RLS, or empty string for public tables.
- All CRUD methods return errors on failure.
- The `Select` method decodes JSON directly into your slice.

## Advanced Query Builder Usage

### Ordering Results
```go
var tenants []Tenant
err := client.Table("tenants").
    OrderBy("created_at", "desc").
    Select(&tenants, jwtToken)
```

### Pagination (Offset & Limit)
```go
var tenants []Tenant
err := client.Table("tenants").
    Offset(10).
    Limit(5).
    Select(&tenants, jwtToken)
```

### Selecting Specific Columns
```go
var tenants []Tenant
err := client.Table("tenants").
    SelectColumns("id", "name", "plan").
    Select(&tenants, jwtToken)
```

### Filtering Examples

#### Equality and Not Equal
```go
// Find all tenants with plan 'pro', but not deleted
var tenants []Tenant
err := client.Table("tenants").
    Eq("plan", "pro").
    NotEq("deleted_at", nil).
    Select(&tenants, jwtToken)
```

#### Handling NULL and nil values
- When using filters like `.Eq("deleted_at", nil)` or `.NotEq("deleted_at", nil)`, the SDK will generate correct `is.null` or `not.is.null` queries.
- The SDK will never send the string `"<nil>"` to Postgres.
- For `IN` filters, nil values in the slice are encoded as `null`.

#### In (matching any value in a slice)
```go
// Find tenants with plan 'pro' or 'enterprise', or where deleted_at is null
var tenants []Tenant
plans := []interface{}{ "pro", "enterprise" }
err := client.Table("tenants").
    In("plan", plans).
    In("deleted_at", []interface{}{nil}).
    Select(&tenants, jwtToken)
```

### Insert: Best Practice for DB Defaults
- Omit fields like `id`, `created_at`, etc. from your struct or set them to `nil`/zero.
- The SDK will omit them from JSON, letting the DB generate values.
- After insert, the returned slice will contain the full DB row(s).

### Error Handling
- All CRUD methods return errors on failure.
- Common Supabase/PostgREST error messages are surfaced directly.

### Compatibility Notes
- `Insert` now supports returning DB-generated fields when passed a pointer to a slice.
- Filters handle nils and pointers correctly; no more invalid timestamp errors.
- Single-record select logic is unchanged; you can still fetch by ID as before.

### Ordering Results
```go
var tenants []Tenant
err := client.Table("tenants").
    OrderBy("created_at", "desc").
    Select(&tenants, jwtToken)
```

### Pagination (Offset & Limit)
```go
var tenants []Tenant
err := client.Table("tenants").
    Offset(10).
    Limit(5).
    Select(&tenants, jwtToken)
```

### Selecting Specific Columns
```go
var tenants []Tenant
err := client.Table("tenants").
    SelectColumns("id", "name", "plan").
    Select(&tenants, jwtToken)
```

### Combined Example
```go
var tenants []Tenant
err := client.Table("tenants").
    Eq("plan", "pro").
    OrderBy("created_at", "desc").
    Offset(10).
    Limit(5).
    SelectColumns("id", "name", "plan").
    Select(&tenants, jwtToken)
```

### Filtering Examples

#### Equality and Not Equal
```go
// Find all tenants with plan 'pro', but not deleted
var tenants []Tenant
err := client.Table("tenants").
    Eq("plan", "pro").
    NotEq("deleted_at", nil).
    Select(&tenants, jwtToken)
```

#### Greater Than, Less Than, GTE, LTE
```go
// Find tenants created after a certain date and with more than 5 users
var tenants []Tenant
err := client.Table("tenants").
    Gt("created_at", "2024-01-01").
    Gte("max_users", 5).
    Select(&tenants, jwtToken)
```

#### Like and ILike (case-insensitive)
```go
// Find tenants whose names start with 'Acme' (case-insensitive)
var tenants []Tenant
err := client.Table("tenants").
    ILike("name", "Acme%") .
    Select(&tenants, jwtToken)
```

#### In (matching any value in a slice)
```go
// Find tenants with plan 'pro' or 'enterprise'
var tenants []Tenant
plans := []interface{}{ "pro", "enterprise" }
err := client.Table("tenants").
    In("plan", plans).
    Select(&tenants, jwtToken)
```

### Combining Filters, Ordering, Pagination, and Column Selection
```go
// Find the first 10 'pro' tenants, ordered by creation date descending, only fetch id and name
var tenants []Tenant
err := client.Table("tenants").
    Eq("plan", "pro").
    OrderBy("created_at", "desc").
    Limit(10).
    SelectColumns("id", "name").
    Select(&tenants, jwtToken)
```

### Full Example: Complex Query
```go
// Find up to 5 tenants created after 2024-01-01, not deleted, whose name contains 'AI',
// ordered by max_users descending, skipping the first 2 results, returning only id, name, and max_users.
var tenants []Tenant
err := client.Table("tenants").
    Gt("created_at", "2024-01-01").
    NotEq("deleted_at", nil).
    ILike("name", "%AI%").
    OrderBy("max_users", "desc").
    Offset(2).
    Limit(5).
    SelectColumns("id", "name", "max_users").
    Select(&tenants, jwtToken)
```

### AND/OR Grouping and Nested Filters

#### Simple OR Group
```go
// Find tenants where plan is 'pro' OR 'enterprise'
var tenants []Tenant
err := client.Table("tenants").
    Or(
        supabasego.Eq("plan", "pro"),
        supabasego.Eq("plan", "enterprise"),
    ).
    Select(&tenants, jwtToken)
```

#### Simple AND Group
```go
// Find tenants where plan is 'pro' AND max_users > 5
var tenants []Tenant
err := client.Table("tenants").
    And(
        supabasego.Eq("plan", "pro"),
        supabasego.Gt("max_users", 5),
    ).
    Select(&tenants, jwtToken)
```

#### Nested AND/OR Group
```go
// Find tenants where (plan = 'pro' AND max_users > 5) OR (plan = 'enterprise' AND max_users > 10)
var tenants []Tenant
err := client.Table("tenants").
    Or(
        supabasego.And(
            supabasego.Eq("plan", "pro"),
            supabasego.Gt("max_users", 5),
        ),
        supabasego.And(
            supabasego.Eq("plan", "enterprise"),
            supabasego.Gt("max_users", 10),
        ),
    ).
    Select(&tenants, jwtToken)
```

#### Combining Grouped Filters with Other Query Builder Features
```go
// Find up to 10 tenants where (plan = 'pro' OR plan = 'enterprise'), not deleted, ordered by creation date
var tenants []Tenant
err := client.Table("tenants").
    Or(
        supabasego.Eq("plan", "pro"),
        supabasego.Eq("plan", "enterprise"),
    ).
    NotEq("deleted_at", nil).
    OrderBy("created_at", "desc").
    Limit(10).
    SelectColumns("id", "name", "plan", "created_at").
    Select(&tenants, jwtToken)
```

#### Deeply Nested Example
```go
// Find tenants where ((plan = 'pro' AND max_users > 5) OR (plan = 'enterprise' AND max_users > 10)) AND name ilike '%AI%'
var tenants []Tenant
err := client.Table("tenants").
    And(
        supabasego.Or(
            supabasego.And(
                supabasego.Eq("plan", "pro"),
                supabasego.Gt("max_users", 5),
            ),
            supabasego.And(
                supabasego.Eq("plan", "enterprise"),
                supabasego.Gt("max_users", 10),
            ),
        ),
        supabasego.ILike("name", "%AI%"),
    ).
    Select(&tenants, jwtToken)
```

> **Tip:** You can freely mix grouped filters with all other query builder features (ordering, offset, limit, column selection, etc.)

---

**More CRUD and query builder examples will be added as implementation progresses.**
