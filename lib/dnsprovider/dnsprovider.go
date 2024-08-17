// package dnsprovider contains logic to manage DNS providers in different providers
package dnsprovider

import "context"

type DNSProvider interface {
	Domain() (string, error)
	FindOrCreateARecord(ctx context.Context, zoneID, recordName, recordContent string) error
}
