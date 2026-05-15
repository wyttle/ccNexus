package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSkipEmptyProxyBodyReturnsOKBeforeTransform(t *testing.T) {
	w := httptest.NewRecorder()

	if !skipEmptyProxyBody(w, []byte{}) {
		t.Fatal("expected empty body to be skipped")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected empty body to return 200, got %d body %q", w.Code, w.Body.String())
	}
}

func TestSkipEmptyProxyBodySkipsWhitespaceOnlyBody(t *testing.T) {
	w := httptest.NewRecorder()

	if !skipEmptyProxyBody(w, []byte("   \n\t  ")) {
		t.Fatal("expected whitespace-only body to be skipped")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected whitespace-only body to return 200, got %d body %q", w.Code, w.Body.String())
	}
}

func TestSkipEmptyProxyBodyAllowsJSONBody(t *testing.T) {
	w := httptest.NewRecorder()

	if skipEmptyProxyBody(w, []byte(`{"messages":[]}`)) {
		t.Fatal("expected non-empty JSON body to continue")
	}
}

func TestApplyEndpointReasoningEffortSkipsEmptyBody(t *testing.T) {
	got := applyEndpointReasoningEffort([]byte{}, "xhigh")
	if len(got) != 0 {
		t.Fatalf("expected empty body to stay empty, got %q", string(got))
	}
}

func TestApplyEndpointReasoningEffortDoesNotCreateInvalidJSONForWhitespaceBody(t *testing.T) {
	body := []byte("   \n\t  ")
	got := applyEndpointReasoningEffort(body, "xhigh")
	if string(got) != string(body) {
		t.Fatalf("expected whitespace body to stay unchanged, got %q", string(got))
	}
}

func TestApplyEndpointReasoningEffortPreservesInvalidJSON(t *testing.T) {
	body := []byte(`{"messages":`)
	got := applyEndpointReasoningEffort(body, "xhigh")
	if string(got) != string(body) {
		t.Fatalf("expected invalid json to stay unchanged, got %q", string(got))
	}
}

func TestApplyEndpointReasoningEffortAddsReasoningAndThinking(t *testing.T) {
	body := []byte(`{"messages":[{"role":"user","content":"hi"}]}`)
	got := applyEndpointReasoningEffort(body, "xhigh")

	var req map[string]interface{}
	if err := json.Unmarshal(got, &req); err != nil {
		t.Fatalf("expected valid json, got error %v and body %q", err, string(got))
	}

	reasoning, ok := req["reasoning"].(map[string]interface{})
	if !ok || reasoning["effort"] != "xhigh" {
		t.Fatalf("expected reasoning.effort xhigh, got %#v", req["reasoning"])
	}

	thinking, ok := req["thinking"].(map[string]interface{})
	if !ok || thinking["type"] != "enabled" || thinking["budget_tokens"] != float64(16384) {
		t.Fatalf("expected thinking enabled with xhigh budget, got %#v", req["thinking"])
	}
}
