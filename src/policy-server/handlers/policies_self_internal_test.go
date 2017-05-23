package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"policy-server/handlers"
	"policy-server/handlers/fakes"

	hfakes "code.cloudfoundry.org/go-db-helpers/fakes"
	"code.cloudfoundry.org/go-db-helpers/testsupport"

	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PoliciesSelfInternal", func() {
	var (
		handler           *handlers.PoliciesSelfInternal
		resp              *httptest.ResponseRecorder
		requestJSON       string
		request           *http.Request
		fakeStore         *fakes.Store
		fakeErrorResponse *fakes.ErrorResponse
		logger            *lagertest.TestLogger
		marshaler         *hfakes.Marshaler
		unmarshaler       *hfakes.Unmarshaler
	)

	BeforeEach(func() {
		var err error

		marshaler = &hfakes.Marshaler{}
		marshaler.MarshalStub = json.Marshal
		unmarshaler = &hfakes.Unmarshaler{}
		unmarshaler.UnmarshalStub = json.Unmarshal
		fakeStore = &fakes.Store{}
		logger = lagertest.NewTestLogger("test")
		fakeErrorResponse = &fakes.ErrorResponse{}
		handler = &handlers.PoliciesSelfInternal{
			Logger:        logger,
			Store:         fakeStore,
			Marshaler:     marshaler,
			Unmarshaler:   unmarshaler,
			ErrorResponse: fakeErrorResponse,
		}
		resp = httptest.NewRecorder()
		requestJSON = `{
			"id": "some-app-guid",
			"protocol": "tcp",
			"port": 8080
	  }`
		request, err = http.NewRequest("POST", "/networking/v0/internal/self_policy", bytes.NewBuffer([]byte(requestJSON)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a policy and returns the tag", func() {
	})

	FContext("when there are errors reading the body bytes", func() {
		BeforeEach(func() {
			request.Body = ioutil.NopCloser(&testsupport.BadReader{})
		})

		It("calls the bad request handler", func() {
			handler.ServeHTTP(resp, request)

			Expect(fakeErrorResponse.BadRequestCallCount()).To(Equal(1))

			w, err, message, description := fakeErrorResponse.BadRequestArgsForCall(0)
			Expect(w).To(Equal(resp))
			Expect(err).To(MatchError("banana"))
			Expect(message).To(Equal("policies-self"))
			Expect(description).To(Equal("failed reading request body"))
		})

	})

	FContext("when there are errors in the request body formatting", func() {
		BeforeEach(func() {
			unmarshaler.UnmarshalReturns(errors.New("banana"))
		})

		It("calls the bad request handler", func() {
			handler.ServeHTTP(resp, request)

			Expect(fakeErrorResponse.BadRequestCallCount()).To(Equal(1))

			w, err, message, description := fakeErrorResponse.BadRequestArgsForCall(0)
			Expect(w).To(Equal(resp))
			Expect(err).To(MatchError("banana"))
			Expect(message).To(Equal("policies-self"))
			Expect(description).To(Equal("invalid values passed to API"))
		})
	})
})
