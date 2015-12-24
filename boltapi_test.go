package boltapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/boltdb/bolt"
	"github.com/marconi/boltapi"
	. "github.com/smartystreets/goconvey/convey"
)

type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

func (r *ResponseRecorder) WriteJson(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = r.Write(b)
	return err
}

func (r *ResponseRecorder) EncodeJson(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func NewRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func TestBucketsEndpoint(t *testing.T) {
	Convey("testing buckets endpoint", t, func() {
		err := exec.Command("rm", "./test.db").Run()
		if err != nil {
			t.Error(err)
		}

		db, err := bolt.Open("./test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			t.Error(err)
		}

		restapi, err := boltapi.NewRestApi(db)
		if err != nil {
			t.Error(err)
		}

		Convey("should be able to add and retrieve bucket", func() {
			request := createRequest("POST", "/api/v1/buckets", map[string]string{"name": "bucket1"}, nil)
			response := NewRecorder()
			restapi.AddBucket(response, request)
			So(response.Code, ShouldEqual, http.StatusOK)

			// existing bucket returns error
			request = createRequest("POST", "/api/v1/buckets", map[string]string{"name": "bucket1"}, nil)
			response = NewRecorder()
			restapi.AddBucket(response, request)
			So(response.Code, ShouldEqual, http.StatusInternalServerError)

			request = createRequest("GET", "/api/v1/buckets/bucket1", nil, map[string]string{"name": "bucket1"})
			response = NewRecorder()
			restapi.GetBucket(response, request)
			So(response.Code, ShouldEqual, http.StatusOK)
			So(response.Body.String(), ShouldEqual, `[]`)
		})

		Convey("should be able to list buckets", func() {
			request := createRequest("GET", "/api/v1/buckets", nil, nil)
			response := NewRecorder()
			restapi.ListBuckets(response, request)
			So(response.Code, ShouldEqual, http.StatusOK)
			So(response.Body.String(), ShouldEqual, `[]`)

			request = createRequest("POST", "/api/v1/buckets", map[string]string{"name": "bucket1"}, nil)
			response = NewRecorder()
			restapi.AddBucket(response, request)
			So(response.Code, ShouldEqual, http.StatusOK)

			request = createRequest("GET", "/api/v1/buckets", nil, nil)
			response = NewRecorder()
			restapi.ListBuckets(response, request)
			So(response.Code, ShouldEqual, http.StatusOK)
			So(response.Body.String(), ShouldEqual, `["bucket1"]`)
		})

		Reset(func() {
			db.Close()
		})
	})
}

// func TestBucketEndpoint(t *testing.T) {

// }

// func TestBucketItemEndpoint(t *testing.T) {

// }

func createRequest(method, urlStr string, body interface{}, pathParams map[string]string) *rest.Request {
	request := test.MakeSimpleRequest(method, urlStr, body)
	return &rest.Request{Request: request, PathParams: pathParams}
}
