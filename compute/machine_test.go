package compute

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/cruciblehq/crux/dirs"
	"github.com/cruciblehq/crux/machines"
	"github.com/cruciblehq/spec/agent"
	"github.com/cruciblehq/spec/trust"
	"github.com/cruciblehq/utils-go/codec"
	"github.com/cruciblehq/utils-go/httpx"
)

type fakeMachine struct {
	mu    sync.Mutex
	key   ed25519.PrivateKey
	cert  *tls.Certificate
	token []byte
	id    trust.MachineID
}

func useStateDirs(t *testing.T) {
	t.Helper()
	t.Cleanup(xdg.Reload)
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	xdg.Reload()
}

func startMachine(t *testing.T, enrollment, token []byte) string {
	t.Helper()
	var e trust.Enrollment
	if err := codec.Decode(enrollment, &e, codec.JSON); err != nil {
		t.Fatal(err)
	}
	root, crl, err := e.Verify()
	if err != nil {
		t.Fatal(err)
	}
	pub, key, _ := ed25519.GenerateKey(rand.Reader)
	self, err := trust.Sign(trust.LeafTemplate(e.Machine.Name(), trust.ValidFor(time.Hour), x509.ExtKeyUsageServerAuth), nil, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	m := &fakeMachine{key: key, cert: &tls.Certificate{Certificate: [][]byte{self.Raw}, PrivateKey: key}, token: token, id: e.Machine}

	mux := http.NewServeMux()
	mux.HandleFunc(agent.EndpointEnroll, m.enroll)
	mux.HandleFunc(agent.EndpointCertify, m.certify)
	mux.HandleFunc(agent.EndpointHandshake, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeHandshakeResponse, agent.HandshakeResponse{Version: agent.Version})
	})
	mux.HandleFunc(agent.EndpointHealth, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, agent.MediaTypeHealthResponse, agent.HealthResponse{})
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{
		Handler:   mux,
		TLSConfig: trust.ServerConfig(m.certificate, root, func() *x509.RevocationList { return crl }),
	}
	go srv.ServeTLS(ln, "", "")
	t.Cleanup(func() { srv.Close() })
	return ln.Addr().String()
}

func (m *fakeMachine) certificate() *tls.Certificate {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cert
}

func (m *fakeMachine) enroll(w http.ResponseWriter, r *http.Request) {
	ekm, err := trust.Exporter(*r.TLS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	request, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{DNSNames: []string{m.id.Name()}}, m.key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := codec.Encode(agent.EnrollResponse{Request: request, Proof: trust.TokenProof(m.token, ekm)}, agent.BodyFormat)
	w.Header().Set(httpx.HeaderContentType, agent.MediaTypeEnrollResponse.With(agent.BodyFormat))
	w.Write(data)
}

func (m *fakeMachine) certify(w http.ResponseWriter, r *http.Request) {
	var req agent.CertifyRequest
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := codec.Decode(data, &req, agent.BodyFormat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.mu.Lock()
	m.cert = &tls.Certificate{Certificate: [][]byte{req.Certificate}, PrivateKey: m.key}
	m.mu.Unlock()
	_ = httpx.Write(w, agent.MediaTypeCertifyResponse, agent.BodyFormat, http.StatusOK, agent.CertifyResponse{})
}

func TestNewMachine(t *testing.T) {
	useStateDirs(t)

	data, err := newMachine("crux")
	if err != nil {
		t.Fatalf("newMachine() = %v, want nil", err)
	}
	var e trust.Enrollment
	if err := codec.Decode(data, &e, codec.JSON); err != nil {
		t.Fatal(err)
	}
	if _, _, err := e.Verify(); err != nil {
		t.Errorf("enrollment Verify() = %v, want nil", err)
	}
	rec, err := machines.Load(dirs.Machines(), "crux")
	if err != nil {
		t.Fatal(err)
	}
	if rec.ID != e.Machine || string(rec.Token) != string(e.Token) {
		t.Error("record does not match the enrollment data")
	}
}

func TestEnrollAndConnect(t *testing.T) {
	useStateDirs(t)
	data, err := newMachine("crux")
	if err != nil {
		t.Fatal(err)
	}
	rec, err := machines.Load(dirs.Machines(), "crux")
	if err != nil {
		t.Fatal(err)
	}
	addr := startMachine(t, data, rec.Token)

	if err := waitForMachine(context.Background(), "crux", addr); err != nil {
		t.Fatalf("waitForMachine() = %v, want nil", err)
	}
	rec, err = machines.Load(dirs.Machines(), "crux")
	if err != nil {
		t.Fatal(err)
	}
	if !rec.Enrolled() {
		t.Error("record Enrolled() = false, want true")
	}

	rt, err := connectMachine(context.Background(), "crux", addr)
	if err != nil {
		t.Fatalf("connectMachine() = %v, want nil", err)
	}
	defer rt.Close()
	if err := rt.Health(context.Background()); err != nil {
		t.Errorf("Health() = %v, want nil", err)
	}
}

func TestEnrollWrongToken(t *testing.T) {
	useStateDirs(t)
	data, err := newMachine("crux")
	if err != nil {
		t.Fatal(err)
	}
	addr := startMachine(t, data, trust.NewToken())

	if err := enrollMachine(context.Background(), "crux", addr); !errors.Is(err, trust.ErrInvalidProof) {
		t.Errorf("enrollMachine() = %v, want %v", err, trust.ErrInvalidProof)
	}
}

func TestConnectNotEnrolled(t *testing.T) {
	useStateDirs(t)
	if _, err := newMachine("crux"); err != nil {
		t.Fatal(err)
	}

	if _, err := connectMachine(context.Background(), "crux", "127.0.0.1:1"); !errors.Is(err, ErrNotEnrolled) {
		t.Errorf("connectMachine() = %v, want %v", err, ErrNotEnrolled)
	}
}

func TestWaitForMachineCancelled(t *testing.T) {
	useStateDirs(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := waitForMachine(ctx, "crux", "127.0.0.1:1"); !errors.Is(err, context.Canceled) {
		t.Errorf("waitForMachine() = %v, want %v", err, context.Canceled)
	}
}
