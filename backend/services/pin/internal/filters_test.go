package pin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinFilters_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      PinFilters
		want    PinFilters
		wantErr bool
	}{
		{
			name: "applies defaults",
			in:   PinFilters{},
			want: PinFilters{SortBy: "created_at", SortOrder: "desc", Limit: 20},
		},
		{
			name: "keeps valid values",
			in:   PinFilters{SortBy: "likes", SortOrder: "asc", Limit: 50, Offset: 10},
			want: PinFilters{SortBy: "likes", SortOrder: "asc", Limit: 50, Offset: 10},
		},
		{
			name: "clamps limit and offset",
			in:   PinFilters{Limit: 500, Offset: -5},
			want: PinFilters{SortBy: "created_at", SortOrder: "desc", Limit: 100, Offset: 0},
		},
		{
			// sort_by is interpolated into SQL, so anything off the whitelist must be rejected
			name:    "rejects unknown sort_by",
			in:      PinFilters{SortBy: "id; DROP TABLE pins"},
			wantErr: true,
		},
		{
			name:    "rejects unknown sort_order",
			in:      PinFilters{SortOrder: "sideways"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.in
			err := f.Validate()

			if tt.wantErr {
				var valErr *errValidation
				require.ErrorAs(t, err, &valErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, f)
		})
	}
}
