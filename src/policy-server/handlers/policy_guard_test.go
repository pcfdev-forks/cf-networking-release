package handlers_test

import (
	"errors"
	"policy-server/handlers"
	"policy-server/handlers/fakes"
	"policy-server/models"
	"policy-server/uaa_client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PolicyGuard", func() {
	var (
		policyGuard   *handlers.PolicyGuard
		fakeCCClient  *fakes.CCClient
		fakeUAAClient *fakes.UAAClient
		tokenData     uaa_client.CheckTokenResponse
		policies      []models.Policy
		spaceGUIDs    []string
		space1        models.Space
		space2        models.Space
		space3        models.Space
	)

	BeforeEach(func() {
		fakeCCClient = &fakes.CCClient{}
		fakeUAAClient = &fakes.UAAClient{}
		policyGuard = &handlers.PolicyGuard{
			CCClient:  fakeCCClient,
			UAAClient: fakeUAAClient,
		}
		policies = []models.Policy{
			{
				Source: models.Source{
					ID: "some-app-guid",
				},
				Destination: models.Destination{
					ID: "some-other-guid",
				},
			},
			{
				Source: models.Source{
					ID: "some-app-guid",
				},
				Destination: models.Destination{
					ID: "yet-another-guid",
				},
			},
		}
		tokenData = uaa_client.CheckTokenResponse{
			Scope:    []string{"network.write"},
			UserID:   "some-developer-guid",
			UserName: "some-developer",
		}
		spaceGUIDs = []string{"space-guid-1", "space-guid-2", "space-guid-3"}
		space1 = models.Space{
			Name:    "space-1",
			OrgGUID: "org-guid-1",
		}
		space2 = models.Space{
			Name:    "space-2",
			OrgGUID: "org-guid-2",
		}
		space3 = models.Space{
			Name:    "space-3",
			OrgGUID: "org-guid-3",
		}

		fakeUAAClient.GetTokenReturns("policy-server-token", nil)
		fakeCCClient.GetSpaceGUIDsReturns(spaceGUIDs, nil)
		fakeCCClient.GetSpaceStub = func(token, spaceGUID string) (*models.Space, error) {
			switch spaceGUID {
			case "space-guid-1":
				{
					return &space1, nil
				}
			case "space-guid-2":
				{
					return &space2, nil
				}
			case "space-guid-3":
				{
					return &space3, nil
				}
			default:
				{
					return nil, errors.New("stub called with unexpected guid")
				}
			}
		}
		fakeCCClient.GetUserSpaceStub = func(token, userGUID string, space models.Space) (*models.Space, error) {
			switch space {
			case space1:
				{
					return &space1, nil
				}
			case space2:
				{
					return &space2, nil
				}
			case space3:
				{
					return &space3, nil
				}
			default:
				{
					return nil, errors.New("stub called with unexpected guid")
				}
			}
		}
	})

	Describe("CheckAccess", func() {

		It("checks that the user can access all apps references in policies", func() {
			authorized, err := policyGuard.CheckAccess(policies, tokenData)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeUAAClient.GetTokenCallCount()).To(Equal(1))
			Expect(fakeCCClient.GetSpaceGUIDsCallCount()).To(Equal(1))
			token, appGUIDs := fakeCCClient.GetSpaceGUIDsArgsForCall(0)
			Expect(token).To(Equal("policy-server-token"))
			Expect(appGUIDs).To(ConsistOf([]string{"some-app-guid", "some-other-guid", "yet-another-guid"}))
			Expect(fakeCCClient.GetSpaceCallCount()).To(Equal(3))
			token, guid := fakeCCClient.GetSpaceArgsForCall(0)
			Expect(token).To(Equal("policy-server-token"))
			Expect(guid).To(Equal("space-guid-1"))
			token, guid = fakeCCClient.GetSpaceArgsForCall(1)
			Expect(token).To(Equal("policy-server-token"))
			Expect(guid).To(Equal("space-guid-2"))
			token, guid = fakeCCClient.GetSpaceArgsForCall(2)
			Expect(token).To(Equal("policy-server-token"))
			Expect(guid).To(Equal("space-guid-3"))
			Expect(fakeCCClient.GetUserSpaceCallCount()).To(Equal(3))
			token, userGUID, checkUserSpace := fakeCCClient.GetUserSpaceArgsForCall(0)
			Expect(token).To(Equal("policy-server-token"))
			Expect(userGUID).To(Equal("some-developer-guid"))
			Expect(checkUserSpace).To(Equal(space1))
			token, userGUID, checkUserSpace = fakeCCClient.GetUserSpaceArgsForCall(1)
			Expect(token).To(Equal("policy-server-token"))
			Expect(userGUID).To(Equal("some-developer-guid"))
			Expect(checkUserSpace).To(Equal(space2))
			token, userGUID, checkUserSpace = fakeCCClient.GetUserSpaceArgsForCall(2)
			Expect(token).To(Equal("policy-server-token"))
			Expect(userGUID).To(Equal("some-developer-guid"))
			Expect(checkUserSpace).To(Equal(space3))
			Expect(authorized).To(BeTrue())
		})

		Context("when the token has network.admin scope", func() {
			BeforeEach(func() {
				tokenData = uaa_client.CheckTokenResponse{
					Scope: []string{"network.admin"},
				}
			})
			It("returns successfully without making extra calls to UAA or CC", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeUAAClient.GetTokenCallCount()).To(Equal(0))
				Expect(fakeCCClient.GetSpaceCallCount()).To(Equal(0))
				Expect(fakeCCClient.GetUserSpaceCallCount()).To(Equal(0))
				Expect(authorized).To(BeTrue())
			})
		})

		Context("when the getting one of the the spaces returns nil", func() {
			BeforeEach(func() {
				fakeCCClient.GetSpaceReturns(nil, nil)
			})
			It("returns false", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).NotTo(HaveOccurred())
				Expect(authorized).To(BeFalse())
			})
		})

		Context("when the getting the users spaces returns nil", func() {
			BeforeEach(func() {
				fakeCCClient.GetUserSpaceReturns(nil, nil)
			})
			It("returns false", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).NotTo(HaveOccurred())
				Expect(authorized).To(BeFalse())
			})
		})

		Context("when the getting the policy server token fails", func() {
			BeforeEach(func() {
				fakeUAAClient.GetTokenReturns("", errors.New("banana"))
			})
			It("returns a useful error", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).To(MatchError("getting token: banana"))
				Expect(authorized).To(BeFalse())
			})
		})

		Context("when the getting the space guids fails", func() {
			BeforeEach(func() {
				fakeCCClient.GetSpaceGUIDsReturns(nil, errors.New("banana"))
			})
			It("returns a useful error", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).To(MatchError("getting space guids: banana"))
				Expect(authorized).To(BeFalse())
			})
		})

		Context("when the getting one of the the spaces fails", func() {
			BeforeEach(func() {
				fakeCCClient.GetSpaceReturns(nil, errors.New("banana"))
			})
			It("returns a useful error", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).To(MatchError("getting space with guid space-guid-1: banana"))
				Expect(authorized).To(BeFalse())
			})
		})

		Context("when the getting the users spaces fails", func() {
			BeforeEach(func() {
				fakeCCClient.GetUserSpaceReturns(nil, errors.New("banana"))
			})
			It("returns a useful error", func() {
				authorized, err := policyGuard.CheckAccess(policies, tokenData)
				Expect(err).To(MatchError("getting space with guid space-guid-1: banana"))
				Expect(authorized).To(BeFalse())
			})
		})
	})
})
