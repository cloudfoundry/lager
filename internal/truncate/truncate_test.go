package truncate_test

import (
	"unsafe"

	"code.cloudfoundry.org/lager/internal/truncate"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const longString = "aaaaaaaaaaaaaaaaaaaaaaaaa"
const expectedTruncatedString = "aaaaaaaa-(truncated)"

var _ = Describe("Truncate", func() {
	Describe("String", func() {
		It("does not truncate at all if maxLength is 0", func() {
			Expect(truncate.String(longString, 0)).To(Equal(longString))
		})
		It("does not truncate strings that are equal to the length limit", func() {
			Expect(truncate.String("foobar", 6)).To(Equal("foobar"))
		})
		It("does not truncate strings that are under the length limit", func() {
			Expect(truncate.String("foobar", 10)).To(Equal("foobar"))
		})
		It("does not truncate any strings under 12 characters long even if the string exceeds maxLength", func() {
			Expect(truncate.String("foobar", 5)).To(Equal("foobar"))
		})
		It("truncates the end of long strings that are above the length limit to the length limit", func() {
			Expect(truncate.String(longString, 20)).To(Equal(expectedTruncatedString))
		})
		Describe("boundary conditions", func() {
			It("truncates when the length limit is exactly 12", func() {
				Expect(truncate.String(longString, 12)).To(Equal("-(truncated)"))
			})
			It("truncates when the length limit is exactly 12+1", func() {
				Expect(truncate.String(longString, 13)).To(Equal("a-(truncated)"))
			})
		})
	})
	Describe("Value", func() {
		Context("strings", func() {
			It("returns a new truncated string", func() {
				v := longString
				outV := truncate.Value(v, 20)
				Expect(outV.(string)).To(Equal(expectedTruncatedString))
			})
			It("leaves the reflect.Value of untruncated strings untouched", func() {
				v := "foobar"
				outV := truncate.Value(v, 20)
				Expect(outV).To(BeIdenticalTo(v))
			})
		})
		Context("non-container values", func() {
			It("leaves values with a type that could not possibly contain a string untouched", func() {
				v := 64
				outV := truncate.Value(v, 20)
				Expect(outV).To(BeIdenticalTo(v))
			})
		})
		Context("structs", func() {
			type dummyStruct struct {
				A string
				B string
			}
			type nestedStruct struct {
				A dummyStruct
				B string
			}
			It("truncates long strings inside of structs", func() {
				v := dummyStruct{A: "foobar", B: longString}
				outV := truncate.Value(v, 20)
				Expect(outV.(dummyStruct)).To(Equal(dummyStruct{A: "foobar", B: expectedTruncatedString}))
			})
			It("leaves structs without long strings untouched", func() {
				v := dummyStruct{A: "foobar", B: "barbaz"}
				outV := truncate.Value(v, 20)
				Expect(outV).To(BeIdenticalTo(v))
			})
			It("truncates long strings inside of nested structs", func() {
				v := nestedStruct{A: dummyStruct{A: "foobar", B: longString}, B: "barbaz"}
				outV := truncate.Value(v, 20)
				Expect(outV.(nestedStruct)).To(Equal(nestedStruct{A: dummyStruct{A: "foobar", B: expectedTruncatedString}, B: "barbaz"}))
			})
			It("leaves nested structs without long strings untouched", func() {
				v := nestedStruct{A: dummyStruct{A: "foobar", B: "bazbaq"}, B: "barbaz"}
				outV := truncate.Value(v, 20)
				Expect(outV).To(BeIdenticalTo(v))
			})
		})
		Context("pointers", func() {
			Context("of strings", func() {
				It("truncates long strings referenced by pointers", func() {
					s := longString
					v := &s
					origPtr := uintptr(unsafe.Pointer(v))
					outV := truncate.Value(v, 20)
					tsp := outV.(*string)
					Expect(uintptr(unsafe.Pointer(tsp))).ToNot(Equal(origPtr))
					Expect(*tsp).To(Equal(expectedTruncatedString))
				})
				It("returns the same pointer if there are no strings truncated", func() {
					s := "foobar"
					v := &s
					origPtr := uintptr(unsafe.Pointer(v))
					outV := truncate.Value(v, 20)
					tsp := outV.(*string)
					Expect(uintptr(unsafe.Pointer(tsp))).To(Equal(origPtr))
				})
			})
			Context("of structs", func() {
				type dummyStruct struct {
					A string
					B string
				}

				It("truncates long strings in structs referenced by pointers", func() {
					v := &dummyStruct{A: "foobar", B: longString}
					origPtr := uintptr(unsafe.Pointer(v))
					outV := truncate.Value(v, 20)
					tsp := outV.(*dummyStruct)
					Expect(uintptr(unsafe.Pointer(tsp))).ToNot(Equal(origPtr))
					Expect(tsp).To(Equal(&dummyStruct{A: "foobar", B: expectedTruncatedString}))
				})
				It("returns the same pointer if there are no strings truncated", func() {
					v := &dummyStruct{A: "foobar", B: "barbaz"}
					origPtr := uintptr(unsafe.Pointer(v))
					outV := truncate.Value(v, 20)
					tsp := outV.(*dummyStruct)
					Expect(uintptr(unsafe.Pointer(tsp))).To(Equal(origPtr))
				})

			})
		})
		Context("maps", func() {
			type dummyStruct struct {
				A string
				B string
			}
			It("truncates long strings inside of maps", func() {
				v := map[string]interface{}{
					"struct": dummyStruct{A: "foobar", B: longString},
					"foo":    longString,
				}
				outV := truncate.Value(v, 20)
				Expect(v).ToNot(Equal(outV))
				Expect(outV.(map[string]interface{})).To(Equal(map[string]interface{}{
					"struct": dummyStruct{A: "foobar", B: expectedTruncatedString},
					"foo":    expectedTruncatedString,
				}))
			})
			It("leaves maps without long strings untouched", func() {
				v := map[string]interface{}{
					"struct": dummyStruct{A: "foobar", B: "bar"},
					"foo":    "bar",
				}
				outV := truncate.Value(v, 20)
				Expect(outV).To(Equal(v))
			})
		})
		Context("arrays", func() {
			It("truncates long strings inside of slices", func() {
				v := [2]string{"foobar", longString}
				outV := truncate.Value(v, 20)
				Expect(v).ToNot(Equal(outV))
				Expect(outV.([2]string)).To(Equal([2]string{"foobar", expectedTruncatedString}))
			})
			It("leaves arrays without long strings untouched", func() {
				v := [2]string{"foobar", "bazbaq"}
				outV := truncate.Value(v, 20)
				Expect(outV).To(Equal(v))
			})
		})
		Context("slices", func() {
			It("truncates long strings inside of slices", func() {
				v := []string{"foobar", longString}
				outV := truncate.Value(v, 20)
				Expect(v).ToNot(Equal(outV))
				Expect(outV.([]string)).To(Equal([]string{"foobar", expectedTruncatedString}))
			})
			It("leaves slices without long strings untouched", func() {
				v := []string{"foobar", "bazbaq"}
				outV := truncate.Value(v, 20)
				Expect(outV).To(Equal(v))
			})
		})
	})
})
