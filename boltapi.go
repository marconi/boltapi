package boltapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/boltdb/bolt"
)

var (
	middlewares = []rest.Middleware{
		&rest.AccessLogApacheMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: true,
		},
		&rest.JsonIndentMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	}

	ErrBucketList        = errors.New("error listing buckets")
	ErrBucketGet         = errors.New("error retrieving bucket")
	ErrBucketMissing     = errors.New("bucket doesn't exist")
	ErrBucketCreate      = errors.New("error creating bucket")
	ErrBucketDelete      = errors.New("error deleting bucket")
	ErrBucketDecodeName  = errors.New("error reading bucket name")
	ErrBucketInvalidName = errors.New("invalid bucket name")
	ErrBucketItemDecode  = errors.New("error reading bucket item")
	ErrBucketItemEncode  = errors.New("error encoding bucket item")
	ErrBucketItemCreate  = errors.New("error creating bucket item")
	ErrBucketItemUpdate  = errors.New("error updating bucket item")
	ErrBucketItemDelete  = errors.New("error deleting bucket item")
)

type ApiError struct {
	customErr error
	origErr   error
}

type BucketItem struct {
	Key   string
	Value interface{}
}

func (item *BucketItem) EncodeKey() []byte {
	return []byte(item.Key)
}

func (item *BucketItem) EncodeValue() ([]byte, error) {
	buf, err := json.Marshal(item.Value)
	if err != nil {
		return nil, ErrBucketItemEncode
	}
	return buf, nil
}

func (item *BucketItem) DecodeValue(rawValue []byte) error {
	if err := json.Unmarshal(rawValue, &item.Value); err != nil {
		return ErrBucketItemDecode
	}
	return nil
}

func (err ApiError) Error() string {
	if err.origErr != nil {
		return fmt.Sprintf("%s: %v", err.customErr, err.origErr)
	}
	return err.customErr.Error()
}

type RestApi struct {
	db  *bolt.DB
	api *rest.Api
}

func NewRestApi(db *bolt.DB) (*RestApi, error) {
	restapi := &RestApi{db: db}

	api := rest.NewApi()
	api.Use(middlewares...)
	router, err := rest.MakeRouter(
		rest.Get("/v1/buckets", restapi.ListBuckets),
		rest.Post("/v1/buckets", restapi.AddBucket),
		rest.Get("/v1/buckets/:name", restapi.GetBucket),
		rest.Delete("/v1/buckets/:name", restapi.DeleteBucket),
		rest.Post("/v1/buckets/:name", restapi.AddBucketItem),
		rest.Get("/v1/buckets/:name/:key", restapi.GetBucketItem),
		rest.Put("/v1/buckets/:name/:key", restapi.UpdateBucketItem),
		rest.Delete("/v1/buckets/:name/:key", restapi.DeleteBucketItem),
	)
	if err != nil {
		return nil, err
	}

	api.SetApp(router)
	restapi.api = api
	return restapi, nil
}

func Serve(db *bolt.DB, port int) error {
	restapi, err := NewRestApi(db)
	if err != nil {
		return err
	}

	http.Handle("/api/", http.StripPrefix("/api", restapi.api.MakeHandler()))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (restapi *RestApi) GetHandler() http.Handler {
	return restapi.api.MakeHandler()
}

func (restapi *RestApi) ListBuckets(w rest.ResponseWriter, r *rest.Request) {
	bucketNames := []string{}
	if err := restapi.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			bucketNames = append(bucketNames, string(name))
			return nil
		})
	}); err != nil {
		log.Println(ApiError{ErrBucketList, err})
		rest.Error(w, ErrBucketList.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(bucketNames)
}

func (restapi *RestApi) AddBucket(w rest.ResponseWriter, r *rest.Request) {
	fail := func(cusromErr, origErr error) {
		log.Println(ApiError{cusromErr, origErr})
		rest.Error(w, cusromErr.Error(), http.StatusInternalServerError)
	}

	payload := make(map[string]string)
	if err := r.DecodeJsonPayload(&payload); err != nil {
		fail(ErrBucketDecodeName, err)
		return
	}

	bucketName, ok := payload["name"]
	if !ok || strings.TrimSpace(bucketName) == "" {
		fail(ErrBucketInvalidName, nil)
		return
	}

	if err := restapi.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(strings.TrimSpace(bucketName)))
		return err
	}); err != nil {
		log.Println(ApiError{ErrBucketCreate, err})
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (restapi *RestApi) GetBucket(w rest.ResponseWriter, r *rest.Request) {
	bucketName := r.PathParam("name")
	keys := []string{}
	if err := restapi.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return ErrBucketMissing
		}

		return bucket.ForEach(func(k, _ []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	}); err != nil {
		log.Println(ApiError{ErrBucketGet, err})
		switch err {
		case ErrBucketGet:
			rest.Error(w, ErrBucketGet.Error(), http.StatusInternalServerError)
		case ErrBucketMissing:
			rest.Error(w, ErrBucketMissing.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteJson(keys)
}

func (restapi *RestApi) DeleteBucket(w rest.ResponseWriter, r *rest.Request) {
	bucketName := r.PathParam("name")
	if err := restapi.db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketName))
	}); err != nil {
		log.Println(ApiError{ErrBucketDelete, err})
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (restapi *RestApi) AddBucketItem(w rest.ResponseWriter, r *rest.Request) {
	fail := func(cusromErr, origErr error) {
		log.Println(ApiError{cusromErr, origErr})
		rest.Error(w, cusromErr.Error(), http.StatusInternalServerError)
	}

	bucketName := r.PathParam("name")
	payload := new(BucketItem)
	if err := r.DecodeJsonPayload(payload); err != nil {
		fail(ErrBucketItemDecode, err)
		return
	}

	encodedValue, err := payload.EncodeValue()
	if err != nil {
		fail(err, nil)
		return
	}

	if err := restapi.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.TrimSpace(bucketName)))
		if bucket == nil {
			return ErrBucketMissing
		}
		return bucket.Put(payload.EncodeKey(), encodedValue)
	}); err != nil {
		fail(ErrBucketItemCreate, err)
		return
	}
}

func (restapi *RestApi) GetBucketItem(w rest.ResponseWriter, r *rest.Request) {
	bucketName := r.PathParam("name")
	bucketItemKey := r.PathParam("key")
	bucketItem := new(BucketItem)
	if err := restapi.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.TrimSpace(bucketName)))
		if bucket == nil {
			return ErrBucketMissing
		}
		itemValue := bucket.Get([]byte(bucketItemKey))
		return bucketItem.DecodeValue(itemValue)
	}); err != nil {
		log.Println(ApiError{err, nil})
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(bucketItem.Value)
}

func (restapi *RestApi) UpdateBucketItem(w rest.ResponseWriter, r *rest.Request) {
	fail := func(cusromErr, origErr error) {
		log.Println(ApiError{cusromErr, origErr})
		rest.Error(w, cusromErr.Error(), http.StatusInternalServerError)
	}

	bucketName := r.PathParam("name")
	bucketItemKey := r.PathParam("key")
	payload := &BucketItem{Key: bucketItemKey}
	if err := r.DecodeJsonPayload(&payload.Value); err != nil {
		fail(ErrBucketItemDecode, err)
		return
	}

	encodedValue, err := payload.EncodeValue()
	if err != nil {
		fail(err, nil)
		return
	}

	if err := restapi.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.TrimSpace(bucketName)))
		if bucket == nil {
			return ErrBucketMissing
		}
		return bucket.Put(payload.EncodeKey(), encodedValue)
	}); err != nil {
		fail(ErrBucketItemUpdate, err)
		return
	}
	w.WriteJson(payload.Value)
}

func (restapi *RestApi) DeleteBucketItem(w rest.ResponseWriter, r *rest.Request) {
	bucketName := r.PathParam("name")
	bucketItemKey := r.PathParam("key")
	if err := restapi.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(strings.TrimSpace(bucketName)))
		if bucket == nil {
			return ErrBucketMissing
		}
		return bucket.Delete([]byte(bucketItemKey))
	}); err != nil {
		log.Println(ApiError{ErrBucketItemDelete, err})
		rest.Error(w, ErrBucketItemDelete.Error(), http.StatusInternalServerError)
		return
	}
}
