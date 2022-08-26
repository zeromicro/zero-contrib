package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var testStrBytes = []byte("Hello World")
var testStrEtag = "11-0a4d55a8d778e5022fab701977c5d840bbc486d0"

type EmptyEtagSuite struct {
	suite.Suite

	server *httptest.Server
}

func (s *EmptyEtagSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.Handle("/", NewETagMiddleware(true).Handle(emptyHandlerFunc))

	s.server = httptest.NewServer(mux)
}

func (s *EmptyEtagSuite) TestNoEtag() {
	res, err := http.Get(s.server.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusNoContent, res.StatusCode)
	s.Empty(res.Header.Get(HeaderETag))
}

func TestEmptyEtag(t *testing.T) {
	suite.Run(t, new(EmptyEtagSuite))
}

type EtagSuite struct {
	suite.Suite

	server     *httptest.Server
	weakServer *httptest.Server
}

func (s *EtagSuite) SetupTest() {
	mux := http.NewServeMux()
	mux.Handle("/", NewETagMiddleware(false).Handle(handlerFunc))

	s.server = httptest.NewServer(mux)

	wmux := http.NewServeMux()
	wmux.Handle("/", NewETagMiddleware(true).Handle(handlerFunc))

	s.weakServer = httptest.NewServer(wmux)
}

func (s EtagSuite) TestEtagExists() {
	res, err := http.Get(s.server.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	h := sha1.New()
	h.Write(testStrBytes)

	s.Equal(fmt.Sprintf("%v-%v", len(testStrBytes), hex.EncodeToString(h.Sum(nil))), res.Header.Get(HeaderETag))
}

func (s EtagSuite) TestWeakEtagExists() {
	res, err := http.Get(s.weakServer.URL + "/")

	s.Nil(err)
	s.Equal(http.StatusOK, res.StatusCode)

	h := sha1.New()
	h.Write(testStrBytes)

	s.Equal(fmt.Sprintf("W/%v-%v", len(testStrBytes), hex.EncodeToString(h.Sum(nil))), res.Header.Get(HeaderETag))
}

func (s EtagSuite) TestMatch() {
	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/", nil)
	s.Nil(err)

	req.Header.Set(HeaderIfNoneMatch, testStrEtag)

	cli := &http.Client{}
	res, err := cli.Do(req)

	s.Nil(err)
	s.Equal(http.StatusNotModified, res.StatusCode)
}

func TestEtag(t *testing.T) {
	suite.Run(t, new(EtagSuite))
}

func emptyHandlerFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusNoContent)

	res.Write(nil)
}

func handlerFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)

	res.Write(testStrBytes)
}

type FreshSuite struct {
	suite.Suite

	reqHeader http.Header
	resHeader http.Header
}

func (s *FreshSuite) SetupTest() {
	s.reqHeader = make(http.Header)
	s.resHeader = make(http.Header)
}

func (s FreshSuite) TestNoCache() {
	s.reqHeader.Set(HeaderCacheControl, "no-cache")
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagEmpty() {
	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagMatch() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.resHeader.Set(HeaderETag, "foo")

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagMismatch() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.resHeader.Set(HeaderETag, "bar")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagMissing() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestWeakEtagMatch() {
	s.reqHeader.Set(HeaderIfNoneMatch, `W/"foo"`)
	s.resHeader.Set(HeaderETag, `W/"foo"`)

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagStrongMatch() {
	s.reqHeader.Set(HeaderIfNoneMatch, `W/"foo"`)
	s.resHeader.Set(HeaderETag, `"foo"`)

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagIfMatch() {
	s.reqHeader.Set(HeaderIfMatch, "foo")
	s.resHeader.Set(HeaderETag, "bar")

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestWeakEtagIfMatch() {
	s.reqHeader.Set(HeaderIfMatch, "W/foo")
	s.resHeader.Set(HeaderETag, "W/bar")

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestStarEtagIfMatch() {
	s.reqHeader.Set(HeaderIfMatch, "*")
	s.resHeader.Set(HeaderETag, "W/bar")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestWeakEtagIfMatchMatched() {
	s.reqHeader.Set(HeaderIfMatch, "W/bar")
	s.resHeader.Set(HeaderETag, "bar")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagIfMatchMatched() {
	s.reqHeader.Set(HeaderIfMatch, "bar")
	s.resHeader.Set(HeaderETag, "bar")

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestStaleOnEtagWeakMatch() {
	s.reqHeader.Set(HeaderIfNoneMatch, `"foo"`)
	s.resHeader.Set(HeaderETag, `W/"foo"`)

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEtagAsterisk() {
	s.reqHeader.Set(HeaderIfNoneMatch, "*")
	s.resHeader.Set(HeaderETag, `"foo"`)

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestModifiedFresh() {
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(4*time.Second))
	s.resHeader.Set(HeaderLastModified, getFormattedTime(2*time.Second))

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestModifiedStale() {
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(2*time.Second))
	s.resHeader.Set(HeaderLastModified, getFormattedTime(4*time.Second))

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestUnmodifiedFresh() {
	s.reqHeader.Set(HeaderIfUnmodifiedSince, getFormattedTime(2*time.Second))
	s.resHeader.Set(HeaderLastModified, getFormattedTime(4*time.Second))

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestUnmodifiedStale() {
	s.reqHeader.Set(HeaderIfUnmodifiedSince, getFormattedTime(4*time.Second))
	s.resHeader.Set(HeaderLastModified, getFormattedTime(2*time.Second))

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestEmptyLastModified() {
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(4*time.Second))

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestBoshAndModifiedFresh() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(4*time.Second))

	s.resHeader.Set(HeaderETag, "bar")
	s.resHeader.Set(HeaderLastModified, getFormattedTime(2*time.Second))

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestBoshAndETagFresh() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(2*time.Second))

	s.resHeader.Set(HeaderETag, "foo")
	s.resHeader.Set(HeaderLastModified, getFormattedTime(4*time.Second))

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestBoshFresh() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(4*time.Second))

	s.resHeader.Set(HeaderETag, "foo")
	s.resHeader.Set(HeaderLastModified, getFormattedTime(2*time.Second))

	s.True(IsFresh(s.reqHeader, s.resHeader))
}

func (s FreshSuite) TestBoshStale() {
	s.reqHeader.Set(HeaderIfNoneMatch, "foo")
	s.reqHeader.Set(HeaderIfModifiedSince, getFormattedTime(2*time.Second))

	s.resHeader.Set(HeaderETag, "bar")
	s.resHeader.Set(HeaderLastModified, getFormattedTime(4*time.Second))

	s.False(IsFresh(s.reqHeader, s.resHeader))
}

func TestFresh(t *testing.T) {
	suite.Run(t, new(FreshSuite))
}

func getFormattedTime(d time.Duration) string {
	return time.Now().Add(d).Format(http.TimeFormat)
}
