package webapp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seaswalker/spring-analysis/internal/deps/springframework/beans"
	"github.com/seaswalker/spring-analysis/internal/testsupport"
	"github.com/seaswalker/spring-analysis/internal/webapp"
)

// server builds the application against the repository's real configuration
// and views, and serves it on a free port.
func server(t *testing.T) *httptest.Server {
	t.Helper()
	root := testsupport.RepoRoot(t)

	saved := beans.ClassPath
	beans.ClassPath = filepath.Join(root, "resources")
	t.Cleanup(func() { beans.ClassPath = saved })

	dispatcher, err := webapp.New(filepath.Join(root, "web"))
	if err != nil {
		t.Fatalf("assembling the application: %v", err)
	}
	srv := httptest.NewServer(dispatcher)
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, srv *httptest.Server, path string) (int, string, string) {
	t.Helper()
	resp, err := http.Get(srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), resp.Header.Get("Content-Type")
}

func post(t *testing.T, srv *httptest.Server, path, contentType, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(srv.URL+path, contentType, strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out)
}

const echoBody = "\n<html>\n    <body>\n        <h1>%s</h1>\n    </body>\n</html>\n"

func echoed(greeting string) string {
	return strings.Replace(echoBody, "%s", greeting, 1)
}

// TestWelcomeFile covers the context root, served from the welcome-file list.
func TestWelcomeFile(t *testing.T) {
	srv := server(t)
	status, body, contentType := get(t, srv, "/spring/")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	want := "<html>\n<body>\n<h2>Hello World!</h2>\n</body>\n</html>\n"
	if body != want {
		t.Errorf("body = %q, want %q", body, want)
	}
	if len(body) != 52 {
		t.Errorf("body is %d bytes, want 52", len(body))
	}
	if contentType != "text/html;charset=UTF-8" {
		t.Errorf("Content-Type = %q", contentType)
	}
}

// TestEcho covers the greeting route, its exact rendered bytes, and the
// UTF-8 round trip.
func TestEcho(t *testing.T) {
	srv := server(t)
	cases := []struct {
		path string
		want string
		size int
	}{
		{"/spring/echo?name=world", echoed("hello world"), 68},
		{"/spring/echo?name=%E4%B8%96%E7%95%8C", echoed("hello 世界"), 69},
		{"/spring/echo?name=", echoed("hello "), 63},
	}
	for _, tc := range cases {
		status, body, contentType := get(t, srv, tc.path)
		if status != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", tc.path, status)
			continue
		}
		if body != tc.want {
			t.Errorf("GET %s: body = %q, want %q", tc.path, body, tc.want)
		}
		if len(body) != tc.size {
			t.Errorf("GET %s: body is %d bytes, want %d", tc.path, len(body), tc.size)
		}
		if contentType != "text/html;charset=UTF-8" {
			t.Errorf("GET %s: Content-Type = %q", tc.path, contentType)
		}
	}
}

// TestEchoAcceptsAnyVerb covers the unqualified @RequestMapping.
func TestEchoAcceptsAnyVerb(t *testing.T) {
	srv := server(t)
	status, body := post(t, srv, "/spring/echo?name=x", "text/plain", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if body != echoed("hello x") {
		t.Errorf("body = %q", body)
	}
}

// TestEchoRequiresName covers the missing required parameter.
func TestEchoRequiresName(t *testing.T) {
	srv := server(t)
	if status, _, _ := get(t, srv, "/spring/echo"); status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
}

// TestEchoAgain covers the JSON route: what binds, what does not, and the
// status each produces.
func TestEchoAgain(t *testing.T) {
	srv := server(t)
	cases := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantBody    string
	}{
		{"age 30", "application/json", `{"name":"bob","age":30}`, 200, echoed("hello bob, your age is 30.")},
		{"absent age", "application/json", `{"name":"bob"}`, 200, echoed("hello bob, your age is null.")},
		// Over the declared maximum, and still the success message: the
		// violation is printed, never added to the binding result.
		{"age 99", "application/json", `{"name":"bob","age":99}`, 200, echoed("hello bob, your age is 99.")},
		{"epoch date", "application/json", `{"name":"bob","age":30,"date":1577934245000}`, 200, echoed("hello bob, your age is 30.")},
		{"DateTimeFormat date", "application/json", `{"name":"bob","age":30,"date":"2020-01-02 03:04:05"}`, 400, ""},
		{"empty body", "application/json", ``, 400, ""},
		{"wrong media type", "text/plain", `{"name":"bob","age":30}`, 415, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := post(t, srv, "/spring/echoAgain", tc.contentType, tc.body)
			if status != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %q)", status, tc.wantStatus, body)
			}
			if tc.wantBody != "" && body != tc.wantBody {
				t.Errorf("body = %q, want %q", body, tc.wantBody)
			}
		})
	}
}

// TestEchoAgainRejectsGet covers the POST-only mapping.
func TestEchoAgainRejectsGet(t *testing.T) {
	srv := server(t)
	if status, _, _ := get(t, srv, "/spring/echoAgain"); status != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", status)
	}
}

// TestUnmappedPaths covers 404s inside and outside the context, and that
// WEB-INF is not public.
func TestUnmappedPaths(t *testing.T) {
	srv := server(t)
	for _, path := range []string{"/spring/missing", "/elsewhere", "/spring/WEB-INF/echo.html"} {
		if status, _, _ := get(t, srv, path); status != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", path, status)
		}
	}
}

// TestStaticResource covers <mvc:default-servlet-handler/>.
func TestStaticResource(t *testing.T) {
	srv := server(t)
	status, body, _ := get(t, srv, "/spring/index.html")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if !strings.Contains(body, "Hello World!") {
		t.Errorf("body = %q", body)
	}
}

// TestConfigurationIsReadFromTheXML checks the assembly really consults
// spring-servlet.xml rather than hard-coding what it says.
func TestConfigurationIsReadFromTheXML(t *testing.T) {
	root := testsupport.RepoRoot(t)
	saved := beans.ClassPath
	beans.ClassPath = filepath.Join(root, "resources")
	defer func() { beans.ClassPath = saved }()

	if _, err := webapp.New(filepath.Join(root, "web")); err != nil {
		t.Fatalf("assembling against the shipped configuration: %v", err)
	}

	beans.ClassPath = t.TempDir()
	if _, err := webapp.New(filepath.Join(root, "web")); err == nil {
		t.Error("assembling with no configuration on the classpath succeeded")
	}
}

// TestDeploymentConstants keep the settings web.xml and the Maven plugin
// fixed.
func TestDeploymentConstants(t *testing.T) {
	if webapp.ContextPath != "/spring" || webapp.DefaultPort != 8080 {
		t.Errorf("deployment settings changed: %q on port %d", webapp.ContextPath, webapp.DefaultPort)
	}
	if webapp.ConfigLocation != "spring-servlet.xml" {
		t.Errorf("config location = %q", webapp.ConfigLocation)
	}
}
