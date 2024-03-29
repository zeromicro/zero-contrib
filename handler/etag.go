package handler

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ETagMiddleware struct {
	weak bool
}

func NewETagMiddleware(weak bool) *ETagMiddleware {
	return &ETagMiddleware{weak: weak}
}

func (m *ETagMiddleware) Handle(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		hw := hashWriter{rw: res, hash: sha1.New(), buf: bytes.NewBuffer(nil)}
		h.ServeHTTP(&hw, req)

		resHeader := res.Header()

		if hw.hash == nil ||
			resHeader.Get(HeaderETag) != "" ||
			strconv.Itoa(hw.status)[0] != '2' ||
			hw.status == http.StatusNoContent ||
			hw.buf.Len() == 0 {
			writeRaw(res, hw)
			return
		}

		etag := fmt.Sprintf("%v-%v", strconv.Itoa(hw.len),
			hex.EncodeToString(hw.hash.Sum(nil)))

		if m.weak {
			etag = "W/" + etag
		}

		resHeader.Set(HeaderETag, etag)

		if IsFresh(req.Header, resHeader) {
			res.WriteHeader(http.StatusNotModified)
			res.Write(nil)
		} else {
			writeRaw(res, hw)
		}
	}
}

type hashWriter struct {
	rw     http.ResponseWriter
	hash   hash.Hash
	buf    *bytes.Buffer
	len    int
	status int
}

func (hw hashWriter) Header() http.Header {
	return hw.rw.Header()
}

func (hw *hashWriter) WriteHeader(status int) {
	hw.status = status
}

func (hw *hashWriter) Write(b []byte) (int, error) {
	if hw.status == 0 {
		hw.status = http.StatusOK
	}
	// bytes.Buffer.Write(b) always return (len(b), nil), so just
	// ignore the return values.
	hw.buf.Write(b)

	l, err := hw.hash.Write(b)
	hw.len += l
	return l, err
}

func writeRaw(res http.ResponseWriter, hw hashWriter) {
	res.WriteHeader(hw.status)
	res.Write(hw.buf.Bytes())
}

// IsFresh check whether cache can be used in this HTTP request
func IsFresh(reqHeader http.Header, resHeader http.Header) bool {
	isEtagMatched, isModifiedMatched := false, false

	ifModifiedSince := reqHeader.Get(HeaderIfModifiedSince)
	ifUnmodifiedSince := reqHeader.Get(HeaderIfUnmodifiedSince)
	ifNoneMatch := reqHeader.Get(HeaderIfNoneMatch)
	ifMatch := reqHeader.Get(HeaderIfMatch)
	cacheControl := reqHeader.Get(HeaderCacheControl)

	etag := resHeader.Get(HeaderETag)
	lastModified := resHeader.Get(HeaderLastModified)

	if ifModifiedSince == "" &&
		ifUnmodifiedSince == "" &&
		ifNoneMatch == "" &&
		ifMatch == "" {
		return false
	}

	if strings.Contains(cacheControl, "no-cache") {
		return false
	}

	if etag != "" && ifNoneMatch != "" {
		isEtagMatched = checkEtagNoneMatch(trimTags(strings.Split(ifNoneMatch, ",")), etag)
	}

	if etag != "" && ifMatch != "" && !isEtagMatched {
		isEtagMatched = checkEtagMatch(trimTags(strings.Split(ifMatch, ",")), etag)
	}

	if lastModified != "" && ifModifiedSince != "" {
		isModifiedMatched = checkModifedMatch(lastModified, ifModifiedSince)
	}

	if lastModified != "" && ifUnmodifiedSince != "" && !isModifiedMatched {
		isModifiedMatched = checkUnmodifedMatch(lastModified, ifUnmodifiedSince)
	}

	return isEtagMatched || isModifiedMatched
}

func trimTags(tags []string) []string {
	trimedTags := make([]string, len(tags))

	for i, tag := range tags {
		trimedTags[i] = strings.TrimSpace(tag)
	}

	return trimedTags
}

func checkEtagNoneMatch(etagsToNoneMatch []string, etag string) bool {
	for _, etagToNoneMatch := range etagsToNoneMatch {
		if etagToNoneMatch == "*" || etagToNoneMatch == etag || etagToNoneMatch == "W/"+etag {
			return true
		}
	}

	return false
}

func checkEtagMatch(etagsToMatch []string, etag string) bool {
	for _, etagToMatch := range etagsToMatch {
		if etagToMatch == "*" {
			return false
		}

		if strings.HasPrefix(etagToMatch, "W/") {
			if etagToMatch == "W/"+etag {
				return false
			}
		} else {
			if etagToMatch == etag {
				return false
			}
		}
	}

	return true
}

func checkModifedMatch(lastModified, ifModifiedSince string) bool {
	if lm, ims, ok := parseTimePairs(lastModified, ifModifiedSince); ok {
		return lm.Before(ims)
	}

	return false
}

func checkUnmodifedMatch(lastModified, ifUnmodifiedSince string) bool {
	if lm, ius, ok := parseTimePairs(lastModified, ifUnmodifiedSince); ok {
		return lm.After(ius)
	}

	return false
}

func parseTimePairs(s1, s2 string) (t1 time.Time, t2 time.Time, ok bool) {
	if t1, err := time.Parse(http.TimeFormat, s1); err == nil {
		if t2, err := time.Parse(http.TimeFormat, s2); err == nil {
			return t1, t2, true
		}
	}

	return t1, t2, false
}
