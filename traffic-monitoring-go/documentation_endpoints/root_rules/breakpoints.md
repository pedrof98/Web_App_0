# Debug Breakpoints Guide for GET /rules Endpoint

## Critical Breakpoints (Set these in order of execution)

### 1. **Rule Handler Entry Point**
**File**: `app/handlers/rule.go`  
**Line**: `func (h *RuleHandler) GetRules(c *gin.Context)`  
**Purpose**: See rule retrieval request and understand the "brain" of the SIEM

### 2. **Filter Parameter Extraction**
**File**: `app/handlers/rule.go`  
**Line**: `status := c.Query("status")`  
**Purpose**: Understand rule filtering by status (enabled/disabled/testing)

### 3. **Category Parameter Extraction**
**File**: `app/handlers/rule.go`  
**Line**: `category := c.Query("category")`  
**Purpose**: See how rules are filtered by category (v2x, network, system, etc.)

### 4. **Query Builder Creation**
**File**: `app/handlers/rule.go`  
**Line**: `query := h.DB.Model(&models.Rule{})`  
**Purpose**: Watch GORM query construction for rule management

### 5. **Status Filter Application**
**File**: `app/handlers/rule.go`  
**Line**: `if status != "" { query = query.Where("status = ?", status) }`  
**Purpose**: See conditional filtering for rule status

### 6. **Category Filter Application**
**File**: `app/handlers/rule.go`  
**Line**: `if category != "" { query = query.Where("category = ?", category) }`  
**Purpose**: Watch category-based rule filtering

### 7. **Rule Ordering**
**File**: `app/handlers/rule.go`  
**Line**: `query = query.Order("name ASC")`  
**Purpose**: Understand rule sorting (alphabetical by name)

### 8. **Database Query Execution**
**File**: `app/handlers/rule.go`  
**Line**: `if err := query.Find(&rules).Error; err != nil {`  
**Purpose**: Watch rule retrieval from database

### 9. **Response Construction**
**File**: `app/handlers/rule.go`  
**Line**: `c.JSON(http.StatusOK, rules)`  
**Purpose**: See final rule list response

## Inspection Variables

At each breakpoint, inspect these key variables:

- `status` - Rule status filter ("enabled", "disabled", "testing")
- `category` - Rule category filter ("v2x", "network", "authentication")
- `query` - GORM query object (watch filter chaining)
- `rules` - Final array of Rule objects
- Rule structure: Name, Description, Condition, Severity, Category, Status
- Rule condition syntax for V2X detection logic