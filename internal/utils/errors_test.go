package utils

import (
	"fmt"
	"testing"
)

func TestErrorManager(t *testing.T) {
	for _, test := range []struct {
		fns  []ManagedErrorFunc
		want error
		desc string
	}{
		{
			fns:  []ManagedErrorFunc{func() error { return nil }},
			want: nil,
			desc: "One function returning nil",
		},
		{
			fns:  []ManagedErrorFunc{func() error { return fmt.Errorf("%s", "error") }},
			want: fmt.Errorf("%s", "error"),
			desc: "One function returning error",
		},
		{
			fns:  []ManagedErrorFunc{func() error { return nil }, func() error { return fmt.Errorf("%s", "error") }},
			want: fmt.Errorf("%s", "error"),
			desc: "One function returning success and one error",
		},
		{
			fns:  []ManagedErrorFunc{func() error { return fmt.Errorf("%s", "error1") }, func() error { return fmt.Errorf("%s", "error2") }},
			want: fmt.Errorf("%s", "error1"),
			desc: "Two functions returning error",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			result := ReturnFirstError(test.fns...)

			if test.want == nil {
				if result != nil {
					t.Errorf("Wanted nil got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Got nil wanted %v", test.want.Error())
				} else if result.Error() != test.want.Error() {
					t.Errorf("Got %v wanted %v", result.Error(), test.want.Error())
				}
			}
		})
	}
}
