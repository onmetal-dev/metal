package billing

import (
	"fmt"
	"strings"

	"github.com/onmetal-dev/metal/lib/store"
)

func UsageHourMeterEventName(offering store.ServerOffering, locationId string) string {
	return strings.ToLower(fmt.Sprintf("%s-%s-%s-usage-hour", offering.ProviderSlug, offering.Id, locationId))
}
