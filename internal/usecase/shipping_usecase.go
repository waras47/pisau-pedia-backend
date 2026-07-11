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
	return u.client.SearchDomesticDestination(ctx, search, 20)
}

// CalculateOptions computes total package weight from the catalog (never
// trusting client-supplied weights) and returns courier quotes for the
// chosen destination.
func (u *ShippingUsecase) CalculateOptions(ctx context.Context, destinationID string, items []OrderItemInput) ([]rajaongkir.ShippingOption, error) {
	if !u.client.Enabled() {
		return nil, ErrShippingUnavailable
	}

	weight := u.totalWeight(ctx, items)
	return u.client.CalculateDomesticCost(ctx, u.client.OriginID(), destinationID, weight, defaultCouriers)
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
