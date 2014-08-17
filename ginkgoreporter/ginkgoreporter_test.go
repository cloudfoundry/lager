package ginkgoreporter_test

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/chug"
	. "github.com/pivotal-golang/lager/ginkgoreporter"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ginkgoreporter", func() {
	var (
		reporter reporters.Reporter
		buffer   *bytes.Buffer
	)

	BeforeEach(func() {
		buffer = &bytes.Buffer{}
		reporter = New(buffer)
	})

	fetchLogs := func() []chug.LogEntry {
		out := make(chan chug.Entry, 1000)
		chug.Chug(buffer, out)
		logs := []chug.LogEntry{}
		for entry := range out {
			if entry.IsLager {
				logs = append(logs, entry.Log)
			}
		}
		return logs
	}

	jsonRoundTrip := func(object interface{}) interface{} {
		jsonEncoded, err := json.Marshal(object)
		Ω(err).ShouldNot(HaveOccurred())
		var out interface{}
		err = json.Unmarshal(jsonEncoded, &out)
		Ω(err).ShouldNot(HaveOccurred())
		return out
	}

	Describe("Announcing the beginning of the suite", func() {
		It("should announce that the suite will begin", func() {
			configType := config.GinkgoConfigType{
				RandomSeed: 1138,
			}
			suiteSummary := &types.SuiteSummary{
				SuiteDescription:           "some description",
				NumberOfSpecsThatWillBeRun: 17,
			}

			reporter.SpecSuiteWillBegin(configType, suiteSummary)
			logs := fetchLogs()[0]
			Ω(logs.LogLevel).Should(Equal(lager.INFO))
			Ω(logs.Source).Should(Equal("ginkgo"))
			Ω(logs.Message).Should(Equal("start-suite"))
			Ω(logs.Session).Should(BeZero())
			Ω(logs.Data["summary"]).Should(Equal(jsonRoundTrip(SuiteStartSummary{
				RandomSeed:                 1138,
				SuiteDescription:           "some description",
				NumberOfSpecsThatWillBeRun: 17,
			})))
		})
	})

	Describe("Announcing the end of the suite", func() {
		var suiteSummary *types.SuiteSummary
		BeforeEach(func() {
			suiteSummary = &types.SuiteSummary{
				SuiteDescription:           "some description",
				NumberOfSpecsThatWillBeRun: 17,
			}
		})

		Context("when the suite succeeds", func() {
			BeforeEach(func() {
				suiteSummary.SuiteSucceeded = true
				suiteSummary.NumberOfPassedSpecs = 17
			})

			It("should info", func() {
				reporter.SpecSuiteDidEnd(suiteSummary)
				logs := fetchLogs()[0]
				Ω(logs.LogLevel).Should(Equal(lager.INFO))
				Ω(logs.Source).Should(Equal("ginkgo"))
				Ω(logs.Message).Should(Equal("end-suite"))
				Ω(logs.Session).Should(BeZero())
				Ω(logs.Data["summary"]).Should(Equal(jsonRoundTrip(SuiteEndSummary{
					SuiteDescription: "some description",
					Passed:           true,
					NumberOfSpecsThatWillBeRun: 17,
					NumberOfPassedSpecs:        17,
					NumberOfFailedSpecs:        0,
				})))
			})
		})

		Context("when the suite fails", func() {
			BeforeEach(func() {
				suiteSummary.SuiteSucceeded = false
				suiteSummary.NumberOfPassedSpecs = 10
				suiteSummary.NumberOfFailedSpecs = 7
			})

			It("should error", func() {
				reporter.SpecSuiteDidEnd(suiteSummary)
				logs := fetchLogs()[0]
				Ω(logs.LogLevel).Should(Equal(lager.ERROR))
				Ω(logs.Source).Should(Equal("ginkgo"))
				Ω(logs.Message).Should(Equal("end-suite"))
				Ω(logs.Error.Error()).Should(Equal("7/17 specs failed"))
				Ω(logs.Session).Should(BeZero())
				Ω(logs.Data["summary"]).Should(Equal(jsonRoundTrip(SuiteEndSummary{
					SuiteDescription: "some description",
					Passed:           false,
					NumberOfSpecsThatWillBeRun: 17,
					NumberOfPassedSpecs:        10,
					NumberOfFailedSpecs:        7,
				})))
			})
		})
	})

	Describe("Announcing specs", func() {
		var summary *types.SpecSummary
		BeforeEach(func() {
			summary = &types.SpecSummary{
				ComponentTexts: []string{"A", "B"},
				ComponentCodeLocations: []types.CodeLocation{
					{
						FileName:       "file/a",
						LineNumber:     3,
						FullStackTrace: "some-stack-trace",
					},
					{
						FileName:       "file/b",
						LineNumber:     4,
						FullStackTrace: "some-stack-trace",
					},
				},
				RunTime: time.Minute,
				State:   types.SpecStatePassed,
			}
		})

		Describe("incrementing sessions", func() {
			It("should increment the session counter as specs run", func() {
				reporter.SpecWillRun(summary)
				reporter.SpecDidComplete(summary)
				reporter.SpecWillRun(summary)
				reporter.SpecDidComplete(summary)

				logs := fetchLogs()
				Ω(logs[0].Session).Should(Equal("1"))
				Ω(logs[1].Session).Should(Equal("1"))
				Ω(logs[2].Session).Should(Equal("2"))
				Ω(logs[3].Session).Should(Equal("2"))
			})
		})

		Context("when a spec starts", func() {
			BeforeEach(func() {
				reporter.SpecWillRun(summary)
			})

			It("should log about the spec starting", func() {
				log := fetchLogs()[0]
				Ω(log.LogLevel).Should(Equal(lager.INFO))
				Ω(log.Source).Should(Equal("ginkgo"))
				Ω(log.Message).Should(Equal("spec.start"))
				Ω(log.Session).Should(Equal("1"))
				Ω(log.Data["summary"]).Should(Equal(jsonRoundTrip(SpecSummary{
					Name:     []string{"A", "B"},
					Location: "file/b:4",
				})))
			})

			Context("when the spec succeeds", func() {
				It("should info", func() {
					reporter.SpecDidComplete(summary)
					log := fetchLogs()[1]
					Ω(log.LogLevel).Should(Equal(lager.INFO))
					Ω(log.Source).Should(Equal("ginkgo"))
					Ω(log.Message).Should(Equal("spec.end"))
					Ω(log.Session).Should(Equal("1"))
					Ω(log.Data["summary"]).Should(Equal(jsonRoundTrip(SpecSummary{
						Name:     []string{"A", "B"},
						Location: "file/b:4",
						State:    "PASSED",
						Passed:   true,
						RunTime:  time.Minute,
					})))
				})
			})

			Context("when the spec fails", func() {
				BeforeEach(func() {
					summary.State = types.SpecStateFailed
					summary.Failure = types.SpecFailure{
						Message: "something failed!",
						Location: types.CodeLocation{
							FileName:       "some/file",
							LineNumber:     3,
							FullStackTrace: "some-stack-trace",
						},
					}
				})

				It("should error", func() {
					reporter.SpecDidComplete(summary)
					log := fetchLogs()[1]
					Ω(log.LogLevel).Should(Equal(lager.ERROR))
					Ω(log.Source).Should(Equal("ginkgo"))
					Ω(log.Message).Should(Equal("spec.end"))
					Ω(log.Session).Should(Equal("1"))
					Ω(log.Error.Error()).Should(Equal("something failed!\nsome/file:3"))
					Ω(log.Data["summary"]).Should(Equal(jsonRoundTrip(SpecSummary{
						Name:       []string{"A", "B"},
						Location:   "file/b:4",
						State:      "FAILED",
						Passed:     false,
						RunTime:    time.Minute,
						StackTrace: "some-stack-trace",
					})))
				})
			})
		})
	})

	Describe("Announcing BeforeSuite and AfterSuite", func() {
		setupSummarySpecs := []struct {
			method  func(reporter reporters.Reporter, setupSummary *types.SetupSummary)
			message string
			name    string
		}{
			{
				reporters.Reporter.BeforeSuiteDidRun,
				"before-suite",
				"BeforeSuite",
			},
			{
				reporters.Reporter.AfterSuiteDidRun,
				"after-suite",
				"AfterSuite",
			},
		}

		for _, setupSummarySpec := range setupSummarySpecs {
			Describe("Announcing "+setupSummarySpec.name, func() {
				var setupSummary *types.SetupSummary

				BeforeEach(func() {
					setupSummary = &types.SetupSummary{
						RunTime:        time.Minute,
						CapturedOutput: "foo",
						SuiteID:        "baz",
					}
				})

				Context("when "+setupSummarySpec.name+" succeeds", func() {
					BeforeEach(func() {
						setupSummary.State = types.SpecStatePassed
					})

					It("should info", func() {
						setupSummarySpec.method(reporter, setupSummary)
						logs := fetchLogs()[0]
						Ω(logs.LogLevel).Should(Equal(lager.INFO))
						Ω(logs.Message).Should(Equal(setupSummarySpec.message))
						Ω(logs.Session).Should(BeZero())
						Ω(logs.Data["summary"]).Should(Equal(jsonRoundTrip(SetupSummary{
							Name:    setupSummarySpec.name,
							State:   "PASSED",
							Passed:  true,
							RunTime: time.Minute,
						})))
					})
				})

				Context("when "+setupSummarySpec.name+" fails", func() {
					BeforeEach(func() {
						setupSummary.State = types.SpecStateFailed
						setupSummary.Failure = types.SpecFailure{
							Message: "something failed!",
							Location: types.CodeLocation{
								FileName:       "some/file",
								LineNumber:     3,
								FullStackTrace: "some-stack-trace",
							},
						}
					})

					It("should error", func() {
						setupSummarySpec.method(reporter, setupSummary)
						logs := fetchLogs()[0]
						Ω(logs.LogLevel).Should(Equal(lager.ERROR))
						Ω(logs.Message).Should(Equal(setupSummarySpec.message))
						Ω(logs.Session).Should(BeZero())
						Ω(logs.Error.Error()).Should(Equal("something failed!\nsome/file:3"))
						Ω(logs.Data["summary"]).Should(Equal(jsonRoundTrip(SetupSummary{
							Name:       setupSummarySpec.name,
							State:      "FAILED",
							Passed:     false,
							RunTime:    time.Minute,
							StackTrace: "some-stack-trace",
						})))
					})
				})
			})
		}
	})
})
