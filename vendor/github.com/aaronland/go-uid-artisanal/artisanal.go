package artisanal

import (
	"context"
	"fmt"
	_ "net/url"
	"strings"
	"sync"

	"github.com/aaronland/go-artisanal-integers/client"
	"github.com/aaronland/go-uid"
)

const ARTISANAL_SCHEME string = "artisanal"

// In principle this could also be done with a sync.OnceFunc call but that will
// require that everyone uses Go 1.21 (whose package import changes broke everything)
// which is literally days old as I write this. So maybe a few releases after 1.21.

var register_mu = new(sync.RWMutex)
var register_map = map[string]bool{}

func init() {

	ctx := context.Background()
	err := RegisterProviderSchemes(ctx)

	if err != nil {
		panic(err)
	}
}

// RegisterProviderSchemes will explicitly register all the schemes associated with the `uid.Provider` interface.
func RegisterProviderSchemes(ctx context.Context) error {

	roster := map[string]uid.ProviderInitializationFunc{}

	for _, s := range client.Schemes() {
		s = strings.Replace(s, "://", "", 1)
		roster[s] = NewArtisanalProvider
	}

	register_mu.Lock()
	defer register_mu.Unlock()

	for scheme, fn := range roster {

		_, exists := register_map[scheme]

		if exists {
			continue
		}

		err := uid.RegisterProvider(ctx, scheme, fn)

		if err != nil {
			return fmt.Errorf("Failed to register client for '%s', %w", scheme, err)
		}

		register_map[scheme] = true
	}

	return nil
}

type ArtisanalProvider struct {
	uid.Provider
	client client.Client
}

type ArtisanalUID struct {
	uid.UID
	id int64
}

func NewArtisanalProvider(ctx context.Context, uri string) (uid.Provider, error) {

	cl_uri := fmt.Sprintf(uri)

	cl, err := client.NewClient(ctx, cl_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create artisanal integer client, %w", err)
	}

	pr := &ArtisanalProvider{
		client: cl,
	}

	return pr, nil
}

func (pr *ArtisanalProvider) UID(ctx context.Context, args ...interface{}) (uid.UID, error) {
	return NewArtisanalUID(ctx, pr.client)
}

func NewArtisanalUID(ctx context.Context, args ...interface{}) (uid.UID, error) {

	if len(args) != 1 {
		return nil, fmt.Errorf("Invalid arguments")
	}

	cl, ok := args[0].(client.Client)

	if !ok {
		return nil, fmt.Errorf("Invalid client")
	}

	i, err := cl.NextInt(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new integerm %w", err)
	}

	u := &ArtisanalUID{
		id: i,
	}

	return u, nil
}

func (u *ArtisanalUID) Value() any {
	return u.id
}

func (u *ArtisanalUID) String() string {
	return fmt.Sprintf("%v", u.Value())
}
