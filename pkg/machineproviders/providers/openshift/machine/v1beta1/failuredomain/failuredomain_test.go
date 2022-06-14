/*
Copyright 2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package failuredomain

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1 "github.com/openshift/api/config/v1"
	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	"github.com/openshift/cluster-control-plane-machine-set-operator/pkg/test/resourcebuilder"
)

var _ = Describe("FailureDomains", func() {
	Context("NewFailureDomains", func() {
		Context("with no failure domains configuration", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				failureDomains, err = NewFailureDomains(machinev1.FailureDomains{})
			})

			It("should not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return a nil list", func() {
				Expect(failureDomains).To(BeNil())
			})
		})

		Context("With AWS failure domain configuration", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				config := resourcebuilder.AWSFailureDomains().BuildFailureDomains()

				failureDomains, err = NewFailureDomains(config)
			})

			It("should not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should construct a list of failure domains", func() {
				Expect(failureDomains).To(ConsistOf(
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1a, Subnet:{Type:id, Value:subenet-us-east-1a}}"),
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1b, Subnet:{Type:id, Value:subenet-us-east-1b}}"),
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1c, Subnet:{Type:id, Value:subenet-us-east-1c}}"),
				))
			})
		})

		Context("With invalid AWS failure domain configuration", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				config := resourcebuilder.AWSFailureDomains().BuildFailureDomains()
				config.AWS = nil

				failureDomains, err = NewFailureDomains(config)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("missing failure domain configuration"))
			})

			It("returns an empty list of failure domains", func() {
				Expect(failureDomains).To(BeEmpty())
			})
		})

		Context("With an unsupported platform type", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				config := machinev1.FailureDomains{
					Platform: configv1.BareMetalPlatformType,
				}

				failureDomains, err = NewFailureDomains(config)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("unsupported platform type: BareMetal"))
			})

			It("returns an empty list of failure domains", func() {
				Expect(failureDomains).To(BeEmpty())
			})
		})
	})

	Context("NewFailureDomainsFromMachines", func() {
		Context("With zero AWS machines", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				failureDomains, err = NewFailureDomainsFromMachines([]machinev1beta1.Machine{}, configv1.AWSPlatformType)
			})

			It("should not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return a empty list", func() {
				Expect(failureDomains).To(BeEmpty())
			})
		})

		Context("With AWS machines", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				providerSpec := resourcebuilder.AWSProviderSpec()
				machines := []machinev1beta1.Machine{}
				for _, az := range []string{"us-east-1a", "us-east-1b", "us-east-1c"} {
					ps := providerSpec.WithAvailabilityZone(az)
					machines = append(machines, *resourcebuilder.Machine().WithProviderSpecBuilder(ps).Build())
				}
				failureDomains, err = NewFailureDomainsFromMachines(machines, configv1.AWSPlatformType)
			})

			It("should not error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should construct a list of failure domains", func() {
				Expect(failureDomains).To(ConsistOf(
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1a, Subnet:{Type:filters, Value:&[{Name:tag:Name Values:[aws-subnet-12345678]}]}}"),
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1b, Subnet:{Type:filters, Value:&[{Name:tag:Name Values:[aws-subnet-12345678]}]}}"),
					HaveField("String()", "AWSFailureDomain{AvailabilityZone:us-east-1c, Subnet:{Type:filters, Value:&[{Name:tag:Name Values:[aws-subnet-12345678]}]}}"),
				))
			})
		})

		Context("With invalid providerSpec", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				providerSpec := resourcebuilder.AWSProviderSpec()
				machine := *resourcebuilder.Machine().WithProviderSpecBuilder(providerSpec).Build()
				machine.Spec.ProviderSpec.Value = nil
				failureDomains, err = NewFailureDomainsFromMachines([]machinev1beta1.Machine{machine}, configv1.AWSPlatformType)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(errMachineMissingProviderSpec))
			})

			It("returns an nil list", func() {
				Expect(failureDomains).To(BeNil())
			})
		})

		Context("With an unsupported platform type", func() {
			var failureDomains []FailureDomain
			var err error

			BeforeEach(func() {
				providerSpec := resourcebuilder.AWSProviderSpec()
				machine := *resourcebuilder.Machine().WithProviderSpecBuilder(providerSpec).Build()
				failureDomains, err = NewFailureDomainsFromMachines([]machinev1beta1.Machine{machine}, configv1.BareMetalPlatformType)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("unsupported platform type: BareMetal")))
			})

			It("returns an nil list", func() {
				Expect(failureDomains).To(BeNil())
			})
		})
	})

	Context("an AWS failure domain", func() {
		var fd failureDomain

		BeforeEach(func() {
			fd = failureDomain{
				platformType: configv1.AWSPlatformType,
			}
		})

		Context("with an availability zone", func() {
			BeforeEach(func() {
				fd.aws = resourcebuilder.AWSFailureDomain().WithAvailabilityZone("us-east-1a").Build()
			})

			It("returns the availability zone for String()", func() {
				Expect(fd.String()).To(Equal("AWSFailureDomain{AvailabilityZone:us-east-1a}"))
			})
		})

		Context("with no availability zone", func() {
			Context("with an ARN type subnet", func() {
				BeforeEach(func() {
					subnetARN := "subnet-us-east-1a"

					fd.aws = resourcebuilder.AWSFailureDomain().WithSubnet(machinev1.AWSResourceReference{
						Type: machinev1.AWSARNReferenceType,
						ARN:  &subnetARN,
					}).Build()
				})

				It("returns the subnet for String()", func() {
					Expect(fd.String()).To(Equal("AWSFailureDomain{Subnet:{Type:arn, Value:subnet-us-east-1a}}"))
				})
			})

			Context("with a filter type subnet", func() {
				BeforeEach(func() {
					fd.aws = resourcebuilder.AWSFailureDomain().WithSubnet(machinev1.AWSResourceReference{
						Type: machinev1.AWSFiltersReferenceType,
						Filters: &[]machinev1.AWSResourceFilter{
							{
								Name:   "tag:Name",
								Values: []string{"subnet-us-east-1b"},
							},
						},
					}).Build()
				})

				It("returns the subnet for String()", func() {
					Expect(fd.String()).To(Equal("AWSFailureDomain{Subnet:{Type:filters, Value:&[{Name:tag:Name Values:[subnet-us-east-1b]}]}}"))
				})
			})

			Context("with an ID type subnet", func() {
				BeforeEach(func() {
					subnetID := "subnet-us-east-1c"

					fd.aws = resourcebuilder.AWSFailureDomain().WithSubnet(machinev1.AWSResourceReference{
						Type: machinev1.AWSIDReferenceType,
						ID:   &subnetID,
					}).Build()
				})

				It("returns the subnet for String()", func() {
					Expect(fd.String()).To(Equal("AWSFailureDomain{Subnet:{Type:id, Value:subnet-us-east-1c}}"))
				})
			})
		})
	})
})
