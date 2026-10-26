package integration

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

const appURL = "http://127.0.0.1:8080"

const down = -1

type fakeExternal struct {
	mu     sync.Mutex
	status map[string]int
	calls  map[string]int
}

func (f *fakeExternal) set(host string, status int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status[host] = status
}

func (f *fakeExternal) count(host string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[host]
}

func (f *fakeExternal) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.calls[r.URL.Hostname()]++
	status := f.status[r.URL.Hostname()]
	f.mu.Unlock()

	if status == down {
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
		return
	}
	w.WriteHeader(status)
}

var client = &http.Client{Transport: &http.Transport{Proxy: nil}}

func startApp(t *testing.T) *fakeExternal {
	bin := os.Getenv("APP_BIN")
	if bin == "" {
		t.Skip("APP_BIN is not set")
	}

	ext := &fakeExternal{status: map[string]int{}, calls: map[string]int{}}
	proxy := httptest.NewServer(ext)
	t.Cleanup(proxy.Close)

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "HTTP_PROXY="+proxy.URL, "http_proxy="+proxy.URL, "NO_PROXY=", "no_proxy=")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })

	for range 50 {
		if conn, err := net.Dial("tcp", "127.0.0.1:8080"); err == nil {
			conn.Close()
			return ext
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("app did not start on :8080")
	return nil
}

type response struct {
	code int
	text string
	json map[string]any
}

func do(t *testing.T, method, path, body string) response {
	t.Helper()
	req, _ := http.NewRequest(method, appURL+path, strings.NewReader(body))
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	r := response{code: resp.StatusCode, text: string(raw)}
	json.Unmarshal(raw, &r.json)
	return r
}

func expect(t *testing.T, r response, code int, text string) {
	t.Helper()
	if r.code != code {
		t.Errorf("want %d, got %d (%q)", code, r.code, r.text)
	}
	if text != "" && !strings.Contains(r.text, text) {
		t.Errorf("want body containing %q, got %q", text, r.text)
	}
}

func TestIntegration(t *testing.T) {
	ext := startApp(t)
	const fraud, billing = "card-fraud.internal", "billing-provider.internal"

	ext.set(fraud, http.StatusForbidden)
	expect(t, do(t, "POST", "/subscriptions", `{"account_id":1,"plan":"pro"}`), 403, "card looks fraudulent")
	expect(t, do(t, "GET", "/subscriptions/1", ""), 404, "")

	ext.set(fraud, down)
	expect(t, do(t, "POST", "/subscriptions", `{"account_id":2,"plan":"basic"}`), 201, `{"id":1,"status":"trial"}`)

	ext.set(fraud, http.StatusOK)
	expect(t, do(t, "POST", "/subscriptions", `{"account_id":1,"plan":"pro"}`), 201, `{"id":2,"status":"trial"}`)
	expect(t, do(t, "POST", "/subscriptions", `{"account_id":1,"plan":"gold"}`), 400, "unknown plan")
	expect(t, do(t, "POST", "/subscriptions", `garbage`), 400, "")

	if n := ext.count(fraud); n != 3 {
		t.Errorf("want 3 antifraud calls, got %d", n)
	}

	r := do(t, "GET", "/subscriptions/1", "")
	expect(t, r, 200, `"AccountID":2,"Plan":"basic","Status":"trial"`)

	expect(t, do(t, "PATCH", "/subscriptions/1/plan", `{"plan":"max"}`), 200, `"Plan":"max","Status":"trial"`)
	expect(t, do(t, "PATCH", "/subscriptions/1/plan", `{"plan":"pro"}`), 409, "cannot downgrade plan with an active debt")
	expect(t, do(t, "PATCH", "/subscriptions/2/plan", `{"plan":"basic"}`), 200, `"Plan":"basic","Status":"trial"`)
	expect(t, do(t, "PATCH", "/subscriptions/2/plan", `{"plan":"gold"}`), 400, "unknown plan")
	expect(t, do(t, "PATCH", "/subscriptions/9/plan", `{"plan":"gold"}`), 400, "unknown plan")
	expect(t, do(t, "PATCH", "/subscriptions/9/plan", `{"plan":"pro"}`), 404, "")
	expect(t, do(t, "PATCH", "/subscriptions/abc/plan", `{"plan":"pro"}`), 400, "")
	expect(t, do(t, "GET", "/subscriptions/abc", ""), 400, "")

	if n := ext.count(billing); n != 0 {
		t.Errorf("billing must not be called for trial subscriptions, got %d calls", n)
	}

	if got := do(t, "GET", "/subscriptions/1", ""); got.json["Plan"] != "max" {
		t.Errorf("plan change not persisted: %q", got.text)
	}
}
