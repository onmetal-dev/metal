package dbstore

import (
	"github.com/onmetal-dev/metal/lib/store"
	"github.com/samber/lo"
)

var hetznerEuroLocations = []store.Location{
	{Id: "NBG1", Continent: "Europe", Country: "Germany", City: "Nuremberg"},
	{Id: "FSN1", Continent: "Europe", Country: "Germany", City: "Falkenstein"},
	{Id: "HEL1", Continent: "Europe", Country: "Finland", City: "Helsinki"},
}

var serverOfferings = []store.ServerOffering{
	{
		Id:           "AX41-NVMe",
		ProviderSlug: store.ProviderSlugHetzner,
		Locations:    hetznerEuroLocations,
		Type:         "AX41-NVMe",
		Description:  "64GB DDR4 RAM, 2x 512GB NVMe SSDs, and AMD Ryzen 5 3600 Hexa-Core \"Matisse\" (Zen2)",
		Cpu: store.Cpu{
			Brand:        "AMD",
			Family:       "AMD Ryzen 3000 Series",
			Name:         "AMD Ryzen 5 3600",
			Cores:        6,
			Threads:      12,
			BaseSpeedGHz: 3.6,
			MaxSpeedGHz:  4.2,
			Arch:         "x86",
		},
		MemoryGB:       64,
		TotalStorageGB: 512 * 2,
		Bandwidth: store.Bandwidth{
			SpeedGbps: 1,
			Unlimited: true,
		},
		Disks: []store.Disk{
			{
				SizeGB: 512,
				Type:   "nvme-ssd",
				Device: "/dev/nvme01n1",
			},
			{
				SizeGB: 512,
				Type:   "nvme-ssd",
				Device: "/dev/nvme02n1",
			},
		},
		Prices: []store.Price{
			{
				LocationId: "HEL1",
				Setup:      0.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.0657,
				Monthly:    41.03,
			},
			{
				LocationId: "FSN1",
				Setup:      0.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.0754,
				Monthly:    47.08,
			},
			{
				LocationId: "NBG1",
				Setup:      0.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.0754,
				Monthly:    47.08,
			},
		},
	},
	{
		Id:           "AX102",
		ProviderSlug: store.ProviderSlugHetzner,
		Locations:    hetznerEuroLocations,
		Type:         "AX102",
		Description:  "128GB ram, 2x 1.92TB NVMe SSDs, and a very fine 16 core chip with gobs of cache",
		Cpu: store.Cpu{
			Brand:        "AMD",
			Family:       "AMD Ryzen 7000 Series",
			Name:         "AMD Ryzen 9 7950X3D",
			Cores:        16,
			Threads:      32,
			BaseSpeedGHz: 4.2,
			MaxSpeedGHz:  5.7,
			Arch:         "x86",
		},
		MemoryGB:       128,
		TotalStorageGB: 1920 * 2,
		Bandwidth: store.Bandwidth{
			SpeedGbps: 1,
			Unlimited: true,
		},
		Disks: []store.Disk{
			{
				SizeGB: 1920,
				Type:   "nvme-ssd",
				Device: "/dev/nvme01n1",
			},
			{
				SizeGB: 1920,
				Type:   "nvme-ssd",
				Device: "/dev/nvme02n1",
			},
		},
		Prices: []store.Price{
			{
				LocationId: "HEL1",
				Setup:      39.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.1832,
				Monthly:    114.40,
			},
			{
				LocationId: "FSN1",
				Setup:      39.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.1921,
				Monthly:    119.90,
			},
			{
				LocationId: "NBG1",
				Setup:      39.00,
				Currency:   store.CurrencyEUR,
				Hourly:     0.1921,
				Monthly:    119.90,
			},
		},
	},
	{
		Id:           "todo",
		ProviderSlug: store.ProviderSlugOVHUS,
		// Location: store.Location{Continent: "Europe", Country: "Germany", City: "Falkenstein"},
		// Location: store.Location{Continent: "Europe", Country: "Germany", City: "Nuremberg"},
		Locations:   []store.Location{},
		Type:        "ADV-02",
		Description: "US datacenter, 2x 960 GB NVMe SSDs, and a battle-tested datacenter CPU",
		Cpu: store.Cpu{
			Brand:        "AMD",
			Family:       "EPYC 4004 Series",
			Name:         "AMD EPYC 4344P",
			Cores:        8,
			Threads:      16,
			BaseSpeedGHz: 3.8,
			MaxSpeedGHz:  5.3,
			Arch:         "x86",
		},
		MemoryGB:       64,
		TotalStorageGB: 960 * 2,
		Bandwidth: store.Bandwidth{
			SpeedGbps: 1,
			Unlimited: true,
		},
		Disks: []store.Disk{
			{
				SizeGB: 960,
				Type:   "nvme-ssd",
				Device: "/dev/nvme01n1",
			},
			{
				SizeGB: 960,
				Type:   "nvme-ssd",
				Device: "/dev/nvme02n1",
			},
		},
		Prices: []store.Price{{
			Setup:    149.00,
			Currency: store.CurrencyUSD,
			Hourly:   149.00 * 12 / (365 * 24),
			Monthly:  149.00,
			Daily:    149.00 * 12 / 365,
		},
		},
	},
}

type ServerOfferingStore struct {
	//db             *gorm.DB
}

var _ store.ServerOfferingStore = ServerOfferingStore{}

type NewServerOfferingStoreParams struct {
	//DB             *gorm.DB
}

func NewServerOfferingStore(params NewServerOfferingStoreParams) *ServerOfferingStore {
	return &ServerOfferingStore{
		//db:             params.DB,
	}
}

func (s ServerOfferingStore) GetServerOfferings() ([]store.ServerOffering, error) {
	return serverOfferings, nil
}

func (s ServerOfferingStore) GetServerOffering(id string) (*store.ServerOffering, error) {
	var offering store.ServerOffering
	offering, ok := lo.Find(serverOfferings, func(o store.ServerOffering) bool {
		return o.Id == id
	})
	if !ok {
		return nil, nil
	}
	return &offering, nil
}
