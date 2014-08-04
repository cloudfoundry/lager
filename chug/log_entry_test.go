package chug_test

import (
	. "github.com/pivotal-golang/lager/chug"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LogEntry", func() {
	Describe("Message", func() {
		It("should reconstitute the message from the source, actions, and tasks", func() {
			entry := LogEntry{
				Source: "source",
				Tasks: []Task{
					{"task-1", "1"},
					{"task-2", "2"},
				},
				Action: "action",
			}

			Ω(entry.Message()).Should(Equal("task-1.task-2.action"))
		})

		Context("with no tasks", func() {
			It("should just have the source and action", func() {
				entry := LogEntry{
					Source: "source",
					Action: "action",
				}

				Ω(entry.Message()).Should(Equal("action"))
			})
		})
	})

	Describe("Session", func() {
		It("should reconstitute the session from the tasks", func() {
			entry := LogEntry{
				Source: "source",
				Tasks: []Task{
					{"task-1", "1"},
					{"task-2", "2"},
				},
				Action: "action",
			}

			Ω(entry.Session()).Should(Equal("1.2"))
		})

		Context("with no tasks", func() {
			It("should be empty", func() {
				entry := LogEntry{
					Source: "source",
					Action: "action",
				}

				Ω(entry.Session()).Should(BeEmpty())
			})
		})
	})
})
