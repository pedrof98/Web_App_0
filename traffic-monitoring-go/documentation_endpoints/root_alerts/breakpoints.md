# Debug Breakpoints Guide for GET /alerts Endpoint

## Critical Breakpoints (Set these in order of execution)

### 1. **Alert Handler Entry Point**
**File**: `app/handlers/alert.go`  
**Line**: `func (h *AlertHandler) GetAlerts(c *gin.Context)`  
**Purpose**: See alert retrieval request and parameter processing

### 2. **Pagination Parameter Extraction**
**File**: `app/handlers/alert.go`  
**Line**: `page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))`  
**Purpose**: Understand pagination logic and parameter defaults

### 3. **Filter Parameter Processing**
**File**: `app/handlers/alert.go`  
**Line**: `severity := c.Query("severity")`  
**Purpose**: See how filtering parameters are extracted and processed

### 4. **Query Builder Creation**
**File**: `app/handlers/alert.go`  
**Line**: `query := h.DB.Model(&models.Alert{}).Preload("Rule").Preload("SecurityEvent").Preload("SecurityEvent.LogSource")`  
**Purpose**: Watch complex GORM query construction with multiple preloads

### 5. **Severity Filter Application**
**File**: `app/handlers/alert.go`  
**Line**: `if severity != "" { query = query.Where("severity = ?", severity) }`  
**Purpose**: See conditional filter application

### 6. **Status Filter Application**
**File**: `app/handlers/alert.go`  
**Line**: `if status != "" { query = query.Where("status = ?", status) }`  
**Purpose**: Watch multiple filter chaining

### 7. **Ordering and Sorting**
**File**: `app/handlers/alert.go`  
**Line**: `query = query.Order("timestamp DESC")`  
**Purpose**: Understand default sorting (most recent first)

### 8. **Total Count Query**
**File**: `app/handlers/alert.go`  
**Line**: `query.Count(&total)`  
**Purpose**: See how pagination total is calculated

### 9. **Paginated Results Query**
**File**: `app/handlers/alert.go`  
**Line**: `if err := query.Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {`  
**Purpose**: Watch paginated data retrieval with all preloaded relationships

### 10. **Response Construction**
**File**: `app/handlers/alert.go`  
**Line**: `c.JSON(http.StatusOK, gin.H{`  
**Purpose**: See final response with data and pagination metadata

## Inspection Variables

At each breakpoint, inspect these key variables:

- `page`, `pageSize`, `offset` - Pagination parameters
- `severity`, `status` - Filter parameters  
- `query` - GORM query object (watch it build up)
- `total` - Total count for pagination
- `alerts` - Final alert array with preloaded relationships
- `alert.Rule` - Associated rule object
- `alert.SecurityEvent` - Associated security event
- `alert.SecurityEvent.LogSource` - Nested log source information
- Response structure with data and pagination