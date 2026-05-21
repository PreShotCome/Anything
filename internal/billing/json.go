package billing

import (
	"encoding/json"
	"io"
)

// jsonDecode is a tiny shim so stripe.go can decode from anything readable —
// *bytes.Buffer in tests, *http.Response.Body in production.
func jsonDecode(r io.Reader, into any) error {
	return json.NewDecoder(r).Decode(into)
}
