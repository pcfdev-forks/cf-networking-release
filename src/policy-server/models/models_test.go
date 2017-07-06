package models_test

import (
	"encoding/json"
	"policy-server/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MarshalJSON", func() {
	var (
		input       []byte
		destination models.Destination
	)
	BeforeEach(func() {
		input = []byte{}
		destination = models.Destination{}
	})

	Describe("Unmarshal", func() {
		Context("when port is not set and ports is set", func() {
			Context("when the range in ports is more than one port", func() {
				BeforeEach(func() {
					input = []byte(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"ports": { "start": 8080, "end": 8090 }
				}`)
				})
				It("unmarshals the policy", func() {
					err := json.Unmarshal(input, &destination)
					Expect(err).NotTo(HaveOccurred())
					Expect(destination).To(Equal(models.Destination{
						ID:       "some-other-app-guid",
						Protocol: "tcp",
						Ports: models.Ports{
							Start: 8080,
							End:   8090,
						},
					}))
				})
			})
			Context("when the range in ports is one port", func() {
				BeforeEach(func() {
					input = []byte(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"ports": { "start": 8080, "end": 8080 }
				}`)
				})
				It("unmarshals the policy and sets the port field", func() {
					err := json.Unmarshal(input, &destination)
					Expect(err).NotTo(HaveOccurred())
					Expect(destination).To(Equal(models.Destination{
						ID:       "some-other-app-guid",
						Protocol: "tcp",
						Port:     8080,
						Ports: models.Ports{
							Start: 8080,
							End:   8080,
						},
					}))
				})

			})
		})
		Context("when port is set and ports is not set", func() {
			BeforeEach(func() {
				input = []byte(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"port":8080
				}`)
			})
			It("unmarshals the policy with the ports field set", func() {
				err := json.Unmarshal(input, &destination)
				Expect(err).NotTo(HaveOccurred())
				Expect(destination).To(Equal(models.Destination{
					ID:       "some-other-app-guid",
					Protocol: "tcp",
					Port:     8080,
					Ports: models.Ports{
						Start: 8080,
						End:   8080,
					},
				}))
			})
		})
		Context("when both port and ports are set", func() {
			Context("when the range in ports is more than one port", func() {
				BeforeEach(func() {
					input = []byte(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"port":8080,
						"ports": { "start": 123, "end": 456 }
				}`)
				})
				It("unmarshals the policy with the ports field set and ignores the port field", func() {
					err := json.Unmarshal(input, &destination)
					Expect(err).To(MatchError("ports and port mismatch"))
				})
			})
			Context("when the range in ports is one port", func() {
				Context("when the range in ports does not match port", func() {
					BeforeEach(func() {
						input = []byte(`{
								"id":"some-other-app-guid",
								"protocol":"tcp",
								"port":123,
								"ports": { "start": 8080, "end": 8080 }
						}`)
					})
					It("unmarshals the policy with the port field set to ports.Start", func() {
						err := json.Unmarshal(input, &destination)
						Expect(err).To(MatchError("ports and port mismatch"))
					})
				})
				Context("when the range in ports matches port", func() {
					BeforeEach(func() {
						input = []byte(`{
							"id":"some-other-app-guid",
							"protocol":"tcp",
							"port":8080,
							"ports": { "start": 8080, "end": 8080 }
						}`)
					})
					It("unmarshals the policy", func() {
						err := json.Unmarshal(input, &destination)
						Expect(err).NotTo(HaveOccurred())
						Expect(destination).To(Equal(
							models.Destination{
								ID:       "some-other-app-guid",
								Protocol: "tcp",
								Port:     8080,
								Ports: models.Ports{
									Start: 8080,
									End:   8080,
								},
							}))
					})
				})
			})
		})
	})

	Describe("Marshal", func() {
		Context("when port is not set and ports is set", func() {
			Context("when the range in ports is more than one port", func() {
				BeforeEach(func() {
					destination = models.Destination{
						ID:       "some-other-app-guid",
						Protocol: "tcp",
						Ports: models.Ports{
							Start: 123,
							End:   456,
						},
					}
				})
				It("marshals the policy", func() {
					marshalled, err := json.Marshal(destination)
					Expect(err).NotTo(HaveOccurred())
					Expect(marshalled).To(MatchJSON(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"ports":{"start":123,"end":456}
					}`))
				})
			})
			Context("when the range in ports is one port", func() {
				BeforeEach(func() {
					destination = models.Destination{
						ID:       "some-other-app-guid",
						Protocol: "tcp",
						Ports: models.Ports{
							Start: 123,
							End:   123,
						},
					}
				})
				It("marshals the policy and sets the port field", func() {
					marshalled, err := json.Marshal(destination)
					Expect(err).NotTo(HaveOccurred())
					Expect(marshalled).To(MatchJSON(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"port": 123,
						"ports":{"start":123,"end":123}
					}`))
				})

			})
		})
		Context("when port is set and ports is not set", func() {
			BeforeEach(func() {
				destination = models.Destination{
					ID:       "some-other-app-guid",
					Protocol: "tcp",
					Port:     8080,
				}
			})
			It("marshals the policy with the ports field set", func() {
				marshalled, err := json.Marshal(destination)
				Expect(err).NotTo(HaveOccurred())
				Expect(marshalled).To(MatchJSON(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"port":8080,
						"ports":{"start":8080,"end":8080}
					}`))
			})
		})
		Context("when both port and ports are set", func() {
			Context("when the range in ports is more than one port", func() {
				BeforeEach(func() {
					destination = models.Destination{
						ID:       "some-other-app-guid",
						Protocol: "tcp",
						Port:     8080,
						Ports: models.Ports{
							Start: 123,
							End:   456,
						},
					}
				})
				It("marshals the policy with the ports field set and ignores the port field", func() {
					_, err := json.Marshal(destination)
					Expect(err.Error()).To(ContainSubstring("json: error calling MarshalJSON for type models.Destination: ports and port mismatch"))
				})
			})
			Context("when the range in ports is one port", func() {
				Context("when the range in ports matches port", func() {
					BeforeEach(func() {
						destination = models.Destination{
							ID:       "some-other-app-guid",
							Protocol: "tcp",
							Port:     123,
							Ports: models.Ports{
								Start: 123,
								End:   123,
							},
						}
					})
					It("marshals the policy", func() {
						marshalled, err := json.Marshal(destination)
						Expect(err).NotTo(HaveOccurred())
						Expect(marshalled).To(MatchJSON(`{
						"id":"some-other-app-guid",
						"protocol":"tcp",
						"port": 123,
						"ports":{"start":123,"end":123}
					}`))
					})
				})
				Context("when the range in ports does not match port", func() {
					BeforeEach(func() {
						destination = models.Destination{
							ID:       "some-other-app-guid",
							Protocol: "tcp",
							Port:     8080,
							Ports: models.Ports{
								Start: 123,
								End:   123,
							},
						}
					})
					It("marshals the policy with the ports field set and ignores the port field", func() {
						_, err := json.Marshal(destination)
						Expect(err.Error()).To(ContainSubstring("json: error calling MarshalJSON for type models.Destination: ports and port mismatch"))
					})
				})
			})
		})
	})
})
