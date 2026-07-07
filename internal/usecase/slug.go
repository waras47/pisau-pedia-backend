package usecase

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var nonSlugChars = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonSlugChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ensureUniqueSlug appends -2, -3, ... to base until existsFn reports no
// conflict, so two products/categories named the same thing never collide
// on the unique slug column.
func ensureUniqueSlug(ctx context.Context, base string, existsFn func(ctx context.Context, slug string) (bool, error)) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := existsFn(ctx, slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}
