package oauth

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/theleeeo/thor/lerror"
)

func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func Test_ParseReturnTo(t *testing.T) {
	testCases := []struct {
		desc           string
		allowedReturns []*url.URL
		r              *http.Request
		want           string
		wantErr        error
	}{
		{
			desc:           "Return is allowed",
			allowedReturns: []*url.URL{mustParseURL("https://theleo.se")},
			r: &http.Request{
				Form: url.Values{
					"return": []string{"https://theleo.se"},
				},
			},
			want: "https://theleo.se",
		},
		{
			desc:           "Return is not allowed",
			allowedReturns: []*url.URL{mustParseURL("https://theleo.se")},
			r: &http.Request{
				Form: url.Values{
					"return": []string{"https://rumpa.nu"},
				},
			},
			wantErr: lerror.New("invalid return url: host is not allowed", http.StatusBadRequest),
		},
		{
			desc:           "No scheme",
			allowedReturns: []*url.URL{mustParseURL("https://theleo.se")},
			r: &http.Request{
				Form: url.Values{
					"return": []string{"theleo.se"},
				},
			},
			wantErr: lerror.New("invalid return url: scheme is missing", http.StatusBadRequest),
		},
		{
			desc:           "Invalid scheme",
			allowedReturns: []*url.URL{mustParseURL("https://theleo.se")},
			r: &http.Request{
				Form: url.Values{
					"return": []string{"http://theleo.se"},
				},
			},
			wantErr: lerror.New("invalid return url: scheme is not allowed", http.StatusBadRequest),
		},
		{
			desc:           "No return url",
			allowedReturns: []*url.URL{mustParseURL("https://theleo.se")},
			r: &http.Request{
				Form: url.Values{},
			},
			want:    "",
			wantErr: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			r, err := parseReturnTo(tC.allowedReturns, tC.r)
			if err != tC.wantErr {
				t.Errorf("parseReturnTo() = %v; want nil", err)
			}
			if r != tC.want {
				t.Errorf("parseReturnTo() = %v; want %v", r, tC.want)
			}
		})
	}
}
