package rates

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/example/currency-converter-api/internal/models"
	"github.com/example/currency-converter-api/internal/providers"
)

type Provider interface {
	Fetch(base string) (map[string]string, time.Time, error)
}

type Service struct {
	db       *gorm.DB
	provider Provider
}

func NewService(db *gorm.DB, p Provider) *Service { return &Service{db: db, provider: p} }

func (s *Service) RefreshAll(base string) error {
	rates, fetchedAt, err := s.provider.Fetch(base)
	if err != nil {
		return err
	}
	b, _ := json.Marshal(rates)
	snap := models.RateSnapshot{
		Provider:  s.Name(),
		Base:      strings.ToUpper(base),
		RatesJSON: string(b),
		FetchedAt: fetchedAt,
	}
	return s.db.Create(&snap).Error
}

func (s *Service) Latest(base string) (map[string]decimal.Decimal, time.Time, error) {
	var snap models.RateSnapshot
	err := s.db.Where("base = ?", strings.ToUpper(base)).Order("created_at DESC").First(&snap).Error
	if err != nil {
		return nil, time.Time{}, err
	}
	m := map[string]string{}
	if err := json.Unmarshal([]byte(snap.RatesJSON), &m); err != nil {
		return nil, time.Time{}, err
	}
	out := map[string]decimal.Decimal{}
	for k, v := range m {
		d, err := decimal.NewFromString(v)
		if err == nil {
			out[k] = d
		}
	}
	return out, snap.FetchedAt, nil
}

func (s *Service) Convert(amount decimal.Decimal, from, to, base string) (decimal.Decimal, error) {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)
	base = strings.ToUpper(base)
	if from == to {
		return amount, nil
	}
	rates, _, err := s.Latest(base)
	if err != nil {
		return decimal.Decimal{}, err
	}
	rTo, okTo := rates[to]
	rFrom, okFrom := rates[from]
	if from == base {
		if !okTo {
			return decimal.Decimal{}, errors.New("Unknown Target Currency")
		}
		return amount.Mul(rTo), nil
	}
	if to == base {
		if !okFrom {
			return decimal.Decimal{}, errors.New("Unknown Source Currency")
		}
		return amount.Div(rFrom), nil
	}
	if !okFrom || !okTo {
		return decimal.Decimal{}, errors.New("Unknown Currency")
	}
	return amount.Mul(rTo).Div(rFrom), nil
}

func (s *Service) Name() string {
	switch s.provider.(type) {
	case *ffProvider:
		return "frankfurter"
	case *erhProvider:
		return "exchangeratehost"
	default:
		return "unknown"
	}
}

type ffProvider struct{}

func (p *ffProvider) Fetch(base string) (map[string]string, time.Time, error) {
	return providers.FetchFrankfurter(base)
}

type erhProvider struct{}

func (p *erhProvider) Fetch(base string) (map[string]string, time.Time, error) {
	return providers.FetchExchangerateHost(base)
}

func NewProvider(name string) Provider {
	switch strings.ToLower(name) {
	case "frankfurter":
		return &ffProvider{}
	case "exchangeratehost":
		return &erhProvider{}
	default:
		return &ffProvider{}
	}
}
