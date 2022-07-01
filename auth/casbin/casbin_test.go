package casbin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
)

func testAuthWithUsernameRequest(t *testing.T, router http.Handler, user string, path string, method string, code int) {
	r, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	request := r.WithContext(context.WithValue(r.Context(), "username", user))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != code {
		t.Errorf("%s, %s, %s: %d, supposed to be %d", user, path, method, w.Code, code)
	}
}
func testDomainAuthWithUsernameRequest(t *testing.T, router http.Handler, user string, domain string, path string, method string, code int) {
	r, _ := http.NewRequestWithContext(context.Background(), method, path, nil)
	ctx := context.WithValue(r.Context(), "username", user)
	request := r.WithContext(context.WithValue(ctx, "domain", domain))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != code {
		t.Errorf("%s, %s,%s, %s: %d, supposed to be %d", user, domain, path, method, w.Code, code)
	}
}

func TestBasic(t *testing.T) {
	e, _ := casbin.NewEnforcer("auth_model.conf", "auth_policy.csv")
	router := NewAuthorizer(e, WithUidField("username"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			_, err := w.Write([]byte("content"))
			assert.Nil(t, err)

			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			flusher.Flush()
		}))
	testAuthWithUsernameRequest(t, router, "alice", "/dataset1/resource1", "GET", 200)
	testAuthWithUsernameRequest(t, router, "alice", "/dataset1/resource1", "POST", 200)
	testAuthWithUsernameRequest(t, router, "alice", "/dataset1/resource2", "GET", 200)
	testAuthWithUsernameRequest(t, router, "alice", "/dataset1/resource2", "POST", 403)
}

func TestBasicDomain(t *testing.T) {
	e, err := casbin.NewEnforcer("auth_model_domain.conf", "auth_policy_domain.csv")
	if err != nil {
		t.Fatal(err)
	}
	router := NewAuthorizer(e, WithUidField("username"), WithDomain("domain"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			_, err := w.Write([]byte("content"))
			assert.Nil(t, err)

			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			flusher.Flush()
		}))
	testDomainAuthWithUsernameRequest(t, router, "alice", "go-zero", "/dataset1/resource1", "POST", 200)
	testDomainAuthWithUsernameRequest(t, router, "bob", "domain1", "/dataset2/resource1", "POST", 200)
	testDomainAuthWithUsernameRequest(t, router, "alice", "go-zero", "/dataset1/resource2", "POST", 200)
}

func TestPathWildcard(t *testing.T) {
	e, _ := casbin.NewEnforcer("auth_model.conf", "auth_policy.csv")
	router := NewAuthorizer(e, WithUidField("username"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			_, err := w.Write([]byte("content"))
			assert.Nil(t, err)

			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			flusher.Flush()
		}))

	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource1", "GET", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource1", "POST", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource1", "DELETE", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource2", "GET", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource2", "POST", 403)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/resource2", "DELETE", 403)

	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item1", "GET", 403)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item1", "POST", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item1", "DELETE", 403)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item2", "GET", 403)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item2", "POST", 200)
	testAuthWithUsernameRequest(t, router, "bob", "/dataset2/folder1/item2", "DELETE", 403)
}

func TestRBAC(t *testing.T) {
	e, _ := casbin.NewEnforcer("auth_model.conf", "auth_policy.csv")
	router := NewAuthorizer(e, WithUidField("username"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			_, err := w.Write([]byte("content"))
			assert.Nil(t, err)

			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			flusher.Flush()
		}))

	// cathy can access all /dataset1/* resources via all methods because it has the dataset1_admin role.
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "GET", 200)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "POST", 200)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "DELETE", 200)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "GET", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "POST", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "DELETE", 403)

	// delete all roles on user cathy, so cathy cannot access any resources now.
	_, err := e.DeleteRolesForUser("cathy")
	if err != nil {
		t.Errorf("got error %v", err)
	}

	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "GET", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "POST", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset1/item", "DELETE", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "GET", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "POST", 403)
	testAuthWithUsernameRequest(t, router, "cathy", "/dataset2/item", "DELETE", 403)
}

func TestUsernameNotFounded(t *testing.T) {
	e, _ := casbin.NewEnforcer("auth_model.conf", "auth_policy.csv")
	router := NewAuthorizer(e, WithUidField("username"))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "test")
			_, err := w.Write([]byte("content"))
			assert.Nil(t, err)

			flusher, ok := w.(http.Flusher)
			assert.True(t, ok)
			flusher.Flush()
		}))

	r, _ := http.NewRequestWithContext(context.Background(), "GET", "/dataset1/resource1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.EqualValues(t, 403, w.Code)
}
