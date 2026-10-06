package pkg

import (
	"strconv"
	"strings"
)

// ToUintFilter parse nilai filter id (uint/int/string) menjadi uint.
// Return ok=false bila nilai tidak valid, agar filter diabaikan.
func ToUintFilter(v any) (uint, bool) {
	switch n := v.(type) {
	case uint:
		return n, true
	case int:
		if n > 0 {
			return uint(n), true
		}
	case int64:
		if n > 0 {
			return uint(n), true
		}
	case float64:
		if n > 0 {
			return uint(n), true
		}
	case string:
		if id, err := strconv.Atoi(strings.TrimSpace(n)); err == nil && id > 0 {
			return uint(id), true
		}
	}
	return 0, false
}

// ToIntParam ambil nilai int dari params map (page/limit),
// kembalikan fallback bila tipe tidak dikenali.
func ToIntParam(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint:
		return int(n)
	case float64:
		return int(n)
	default:
		return fallback
	}
}

// TotalPages hitung jumlah halaman dari total data dan limit per halaman.
func TotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	if total <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}

// ToSortParam validasi kolom dan arah sort agar aman dari SQL injection.
// Kembalikan defaultSort bila input kosong atau kolom tidak diizinkan.
func ToSortParam(v any, defaultSort string, allowedCols []string) string {
	allowed := make(map[string]bool, len(allowedCols))
	for _, col := range allowedCols {
		allowed[col] = true
	}
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultSort
	}
	parts := strings.Fields(s)
	col := strings.ToLower(parts[0])
	if !allowed[col] {
		return defaultSort
	}
	dir := "desc"
	if len(parts) > 1 && strings.ToLower(parts[1]) == "asc" {
		dir = "asc"
	}
	return col + " " + dir
}
