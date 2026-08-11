package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrSKUAlreadyExists = errors.New("sku already exists")

var (
	mmSizeRe    = regexp.MustCompile(`(\d+)\s*mm`)
	parenAbbrRe = regexp.MustCompile(`\(([A-Za-z0-9]+)\)\s*$`)
)

// skuCategoryCode returns a short uppercase code for a category slug, e.g.
// "gyuto" -> "GYU". Falls back to "GEN" (generic) if the slug is empty.
func skuCategoryCode(categorySlug string) string {
	s := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(categorySlug), "-", ""))
	if s == "" {
		return "GEN"
	}
	if len(s) > 3 {
		s = s[:3]
	}
	return s
}

// skuMaterialCode derives a short material code from a "Steel Type" (or
// "Material") product spec, e.g. "High Speed Steel (HSS)" -> "HSS",
// "Carbon steel" -> "CS". Falls back to "STD" if no matching spec exists.
func skuMaterialCode(specs []ProductSpecInput) string {
	for _, s := range specs {
		label := strings.ToLower(strings.TrimSpace(s.Label))
		if label != "steel type" && label != "material" {
			continue
		}
		if code := materialAbbreviation(s.Value); code != "" {
			return code
		}
	}
	return "STD"
}

// maxMaterialInitials caps how many letters materialAbbreviation takes from
// a verbose spec value (e.g. "Carbon Steel ex Carl Schlieper Wooden Saw")
// so the SKU segment stays short and readable instead of spelling out every
// word in an unusually long spec.
const maxMaterialInitials = 4

func materialAbbreviation(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if m := parenAbbrRe.FindStringSubmatch(value); m != nil {
		return strings.ToUpper(m[1])
	}
	var initials strings.Builder
	for _, w := range strings.Fields(value) {
		if initials.Len() >= maxMaterialInitials {
			break
		}
		w = strings.TrimFunc(w, func(r rune) bool { return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) })
		if w == "" {
			continue
		}
		initials.WriteByte(strings.ToUpper(w)[0])
	}
	return initials.String()
}

// skuSizeCode extracts a blade-length-in-mm code from the product name
// (e.g. "Gyuto Nashiji Finish 210mm" -> "210"), falling back to a
// "Blade Length" spec, then "000" if neither is present.
func skuSizeCode(name string, specs []ProductSpecInput) string {
	if m := mmSizeRe.FindStringSubmatch(name); m != nil {
		return m[1]
	}
	for _, s := range specs {
		if strings.EqualFold(strings.TrimSpace(s.Label), "blade length") {
			if m := mmSizeRe.FindStringSubmatch(s.Value); m != nil {
				return m[1]
			}
		}
	}
	return "000"
}

// ensureUniqueSKU appends -001, -002, ... to base until existsFn reports no
// conflict, mirroring ensureUniqueSlug.
func ensureUniqueSKU(ctx context.Context, base string, existsFn func(ctx context.Context, sku string) (bool, error)) (string, error) {
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%03d", base, i)
		exists, err := existsFn(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
}

// GenerateSKU builds a [Kategori]-[Material]-[Ukuran]-[Urutan] SKU for a
// product (e.g. "GYU-HSS-210-001") and guarantees it's unique via existsFn.
func GenerateSKU(ctx context.Context, categorySlug, name string, specs []ProductSpecInput, existsFn func(ctx context.Context, sku string) (bool, error)) (string, error) {
	base := fmt.Sprintf("%s-%s-%s", skuCategoryCode(categorySlug), skuMaterialCode(specs), skuSizeCode(name, specs))
	return ensureUniqueSKU(ctx, base, existsFn)
}
