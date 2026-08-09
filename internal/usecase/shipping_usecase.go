package usecase

import (
	"context"
	"errors"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/repository"
	"github.com/pisaupediaprojek/pisau-pedia-backend/pkg/rajaongkir"
)

var ErrShippingUnavailable = errors.New("shipping service is not configured")

// defaultCouriers is the domestic courier set queried for cost options.
var defaultCouriers = []string{"jne", "sicepat", "jnt", "tiki", "pos", "ninja", "ide", "sap"}

// internationalCouriers is the international courier set.
var internationalCouriers = []string{"pos", "tiki", "jne"}

// minWeightGrams guards against RajaOngkir rejecting a zero/too-small weight.
const minWeightGrams = 1000

type ShippingUsecase struct {
	client      *rajaongkir.Client
	productRepo repository.ProductRepository
}

func NewShippingUsecase(client *rajaongkir.Client, productRepo repository.ProductRepository) *ShippingUsecase {
	return &ShippingUsecase{client: client, productRepo: productRepo}
}

func (u *ShippingUsecase) SearchDestinations(ctx context.Context, search string) ([]rajaongkir.Destination, error) {
	if !u.client.Enabled() {
		return nil, ErrShippingUnavailable
	}
	domestic, err := u.client.SearchDomesticDestination(ctx, search, 20)
	if err != nil {
		domestic = nil
	}

	international, _ := u.client.SearchInternationalDestination(ctx, search, 10)
	return append(domestic, international...), nil
}

// CalculateOptions computes total package weight from the catalog (never
// trusting client-supplied weights) and returns courier quotes for the
// chosen destination.
func (u *ShippingUsecase) CalculateOptions(ctx context.Context, destinationID string, items []OrderItemInput) ([]rajaongkir.ShippingOption, error) {
	if !u.client.Enabled() {
		return nil, ErrShippingUnavailable
	}

	weight := u.totalWeight(ctx, items)

	// Try domestic first
	opts, err := u.client.CalculateDomesticCost(ctx, u.client.OriginID(), destinationID, weight, defaultCouriers)
	if err == nil && len(opts) > 0 {
		return opts, nil
	}

	// Fall back to international
	intlOpts, intlErr := u.client.CalculateInternationalCost(ctx, u.client.OriginID(), destinationID, weight, internationalCouriers)
	if intlErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, intlErr
	}
	return intlOpts, nil
}

func (u *ShippingUsecase) totalWeight(ctx context.Context, items []OrderItemInput) int {
	var grams int
	for _, i := range items {
		if i.Quantity == 0 {
			continue
		}
		product, err := u.productRepo.FindBySlug(ctx, i.ProductSlug)
		if err != nil {
			continue
		}
		grams += int(product.Weight) * int(i.Quantity)
	}
	if grams < minWeightGrams {
		grams = minWeightGrams
	}
	return grams
}
