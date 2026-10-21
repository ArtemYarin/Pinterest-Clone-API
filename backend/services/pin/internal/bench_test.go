package pin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// stubPinRepository serves a fixed page of pins. Benchmarks use stubs instead of
// testify mocks, whose call bookkeeping would dominate the measurements.
type stubPinRepository struct {
	PinRepository // unimplemented methods panic if called
	pins          []PinResponse
}

func (r *stubPinRepository) GetPins(ctx context.Context, filters PinFilters) ([]PinResponse, int, error) {
	return r.pins, len(r.pins), nil
}

// newPresignOnlyStorage returns an ImageStorage that can sign URLs without a
// running server: presigning is local HMAC work when the region is preset.
func newPresignOnlyStorage(b *testing.B) *ImageStorage {
	b.Helper()
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("access-key", "secret-key", ""),
		Region: "garage",
	})
	if err != nil {
		b.Fatal(err)
	}
	return &ImageStorage{client: client, presignClient: client, bucket: "pins"}
}

func makeTestPins(n int) []PinResponse {
	owner := uuid.New()
	desc := "A pin description long enough to resemble real user input."
	pins := make([]PinResponse, n)
	for i := range pins {
		pins[i] = PinResponse{
			Id:           uuid.New(),
			User_id:      owner,
			Title:        fmt.Sprintf("Pin %d", i),
			Image_url:    fmt.Sprintf("pins/%s/%s", owner, uuid.New()),
			Image_status: "confirmed",
			Description:  &desc,
			Created_at:   time.Now(),
			Updated_at:   time.Now(),
			Likes:        i,
		}
	}
	return pins
}

func BenchmarkParseFilters_Validate(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet,
		"/?user_id=6f1c&search=sunset&sort_by=likes&sort_order=asc&limit=50&offset=100", nil)
	b.ReportAllocs()

	for b.Loop() {
		f := ParseFilters(req)
		if err := f.Validate(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPinService_GetPins measures the per-page cost of signing a
// download URL for every pin - the service's hot path for the feed.
func BenchmarkPinService_GetPins(b *testing.B) {
	storage := newPresignOnlyStorage(b)
	filters := PinFilters{SortBy: "created_at", SortOrder: "desc", Limit: 100}

	for _, n := range []int{20, 100} {
		b.Run(fmt.Sprintf("pins=%d", n), func(b *testing.B) {
			svc := NewPinService(&stubPinRepository{pins: makeTestPins(n)}, nil, storage)
			ctx := context.Background()
			b.ReportAllocs()

			for b.Loop() {
				if _, _, err := svc.GetPins(ctx, filters); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	pins := makeTestPins(20)
	page := make([]DownloadImgPinResponse, len(pins))
	for i, p := range pins {
		page[i] = DownloadImgPinResponse{Pin: p, Download_url: "https://download.example/" + p.Image_url}
	}
	b.ReportAllocs()

	for b.Loop() {
		WriteJSON(httptest.NewRecorder(), http.StatusOK, page)
	}
}
