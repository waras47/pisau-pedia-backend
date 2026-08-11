// backfill-sku generates a SKU ([Kategori]-[Material]-[Ukuran]-[Urutan],
// see internal/usecase/sku.go) for every existing product that doesn't have
// one yet. New products get one automatically on create; this is only
// needed once, for products that existed before the sku column was added.
//
// Usage:
//
//	go run ./cmd/backfill-sku
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/infrastructure/mysql"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/config"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	db, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	productRepo := mysql.NewProductRepository(db)
	categoryRepo := mysql.NewCategoryRepository(db)
	ctx := context.Background()

	var products []entity.Product
	if err := db.SelectContext(ctx, &products, `SELECT * FROM products WHERE sku IS NULL ORDER BY created_at`); err != nil {
		log.Fatalf("list products: %v", err)
	}

	fmt.Printf("Found %d product(s) without an SKU.\n\n", len(products))
	if len(products) == 0 {
		return
	}

	categorySlugCache := map[string]string{}
	ok, failed := 0, 0

	for i := range products {
		p := &products[i]

		var specs []entity.ProductSpec
		if err := db.SelectContext(ctx, &specs, `SELECT * FROM product_specs WHERE product_id = ?`, p.ID); err != nil {
			log.Printf("SKIP %q: load specs: %v", p.Name, err)
			failed++
			continue
		}
		specsInput := make([]usecase.ProductSpecInput, len(specs))
		for j, s := range specs {
			specsInput[j] = usecase.ProductSpecInput{Label: s.Label, Value: s.Value}
		}

		categorySlug := ""
		if p.CategoryID != nil {
			slug, cached := categorySlugCache[*p.CategoryID]
			if !cached {
				if cat, err := categoryRepo.FindByID(ctx, *p.CategoryID); err == nil && cat != nil {
					slug = cat.Slug
				}
				categorySlugCache[*p.CategoryID] = slug
			}
			categorySlug = slug
		}

		existsFn := func(ctx context.Context, sku string) (bool, error) {
			return productRepo.ExistsBySKU(ctx, sku, p.ID)
		}
		sku, err := usecase.GenerateSKU(ctx, categorySlug, p.Name, specsInput, existsFn)
		if err != nil {
			log.Printf("FAILED %q: generate sku: %v", p.Name, err)
			failed++
			continue
		}

		if _, err := db.ExecContext(ctx, `UPDATE products SET sku = ? WHERE id = ?`, sku, p.ID); err != nil {
			log.Printf("FAILED %q: save sku: %v", p.Name, err)
			failed++
			continue
		}

		fmt.Printf("%-55s -> %s\n", p.Name, sku)
		ok++
	}

	fmt.Printf("\nDone. %d updated, %d failed.\n", ok, failed)
}
